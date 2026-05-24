import { describe, it, expect, vi, beforeEach } from "vitest";

const {
  mockExecFileAsync,
  mockAccess,
  mockReadFile,
  mockShowWarningMessage,
} = vi.hoisted(() => ({
  mockExecFileAsync: vi.fn(
    async (): Promise<{ stdout: string; stderr: string }> => ({
      stdout: "{}",
      stderr: "",
    }),
  ),
  mockAccess: vi.fn(async () => {}),
  mockReadFile: vi.fn(async (): Promise<string> => {
    throw new Error("ENOENT");
  }),
  mockShowWarningMessage: vi.fn(async () => undefined),
}));

vi.mock("vscode", () => ({
  workspace: {
    workspaceFolders: [],
    getConfiguration: () => ({ get: () => undefined }),
  },
  Uri: { file: (p: string) => ({ fsPath: p }) },
  window: { showWarningMessage: mockShowWarningMessage },
  EventEmitter: class {
    private listeners: Array<() => void> = [];
    event = (listener: () => void) => {
      this.listeners.push(listener);
      return { dispose: () => {} };
    };
    fire = () => {
      for (const l of this.listeners) l();
    };
    dispose = () => {};
  },
}));

vi.mock("node:fs/promises", () => ({
  access: mockAccess,
  readFile: mockReadFile,
}));

vi.mock("node:child_process", async () => {
  const util = await import("node:util");
  const execFileFn: Record<symbol, unknown> = vi.fn() as never;
  execFileFn[util.promisify.custom] = mockExecFileAsync;
  return { execFile: execFileFn };
});

vi.mock("./cli.js", () => ({
  buildCliCommand: async () => ({ bin: "go", args: ["run", "discover"] }),
  formatCliCommand: () => "go run discover",
}));

import { DiscoveryCache, DiscoveryService } from "./discovery.js";

function makeOutputChannel() {
  return {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    debug: vi.fn(),
    show: vi.fn(),
  } as unknown as import("vscode").LogOutputChannel;
}

function successResponse(pkgs: Array<{ importPath: string; dir: string }>) {
  return {
    stdout: JSON.stringify({
      packages: pkgs.map((p) => ({
        importPath: p.importPath,
        dir: p.dir,
        suites: [],
      })),
    }),
    stderr: "",
  };
}

describe("DiscoveryService retry", () => {
  let cache: DiscoveryCache;
  let outputChannel: ReturnType<typeof makeOutputChannel>;
  let service: DiscoveryService;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    mockAccess.mockResolvedValue(undefined);
    mockReadFile.mockRejectedValue(new Error("ENOENT"));
    cache = new DiscoveryCache();
    outputChannel = makeOutputChannel();
    service = new DiscoveryService(
      cache,
      outputChannel as unknown as import("vscode").LogOutputChannel,
    );
  });

  it("succeeds on first attempt without retry", async () => {
    mockExecFileAsync.mockResolvedValueOnce(
      successResponse([{ importPath: "example.com/pkg", dir: "/ws/pkg" }]),
    );

    await service.discover("/ws", ["./..."]);

    expect(mockExecFileAsync).toHaveBeenCalledTimes(1);
    expect(cache.packages).toHaveLength(1);
    expect(outputChannel.debug).not.toHaveBeenCalled();
    expect(mockShowWarningMessage).not.toHaveBeenCalled();
  });

  it("retries on transient failure and succeeds on second attempt", async () => {
    mockExecFileAsync
      .mockRejectedValueOnce(new Error("cannot find package"))
      .mockResolvedValueOnce(
        successResponse([{ importPath: "example.com/pkg", dir: "/ws/pkg" }]),
      );

    const p = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await p;

    expect(mockExecFileAsync).toHaveBeenCalledTimes(2);
    expect(cache.packages).toHaveLength(1);
    expect(outputChannel.debug).toHaveBeenCalledWith(
      expect.stringContaining("attempt 1/3 failed"),
    );
    expect(mockShowWarningMessage).not.toHaveBeenCalled();
  });

  it("retries twice and succeeds on third attempt", async () => {
    mockExecFileAsync
      .mockRejectedValueOnce(new Error("fail 1"))
      .mockRejectedValueOnce(new Error("fail 2"))
      .mockResolvedValueOnce(
        successResponse([{ importPath: "example.com/pkg", dir: "/ws/pkg" }]),
      );

    const p = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p;

    expect(mockExecFileAsync).toHaveBeenCalledTimes(3);
    expect(cache.packages).toHaveLength(1);
    expect(outputChannel.debug).toHaveBeenCalledTimes(2);
    expect(mockShowWarningMessage).not.toHaveBeenCalled();
  });

  it("shows toast only after all retries exhausted", async () => {
    mockExecFileAsync.mockRejectedValue(new Error("persistent failure"));

    const p = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p;

    expect(mockExecFileAsync).toHaveBeenCalledTimes(3);
    expect(cache.packages).toHaveLength(0);
    expect(outputChannel.error).toHaveBeenCalledWith(
      expect.stringContaining("failed after 3 attempts"),
    );
    expect(mockShowWarningMessage).toHaveBeenCalledTimes(1);
  });

  it("does not show duplicate toast on repeated total failures", async () => {
    mockExecFileAsync.mockRejectedValue(new Error("persistent failure"));

    const p1 = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p1;

    const p2 = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p2;

    expect(mockShowWarningMessage).toHaveBeenCalledTimes(1);
  });

  it("resets hasShownError after a successful discovery", async () => {
    mockExecFileAsync.mockRejectedValue(new Error("fail"));

    const p1 = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p1;
    expect(mockShowWarningMessage).toHaveBeenCalledTimes(1);

    mockExecFileAsync.mockResolvedValue(
      successResponse([{ importPath: "example.com/pkg", dir: "/ws/pkg" }]),
    );

    await service.discover("/ws", ["./..."]);

    mockExecFileAsync.mockRejectedValue(new Error("fail again"));

    const p3 = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p3;

    expect(mockShowWarningMessage).toHaveBeenCalledTimes(2);
  });

  it("logs retries at debug level, final failure at error level", async () => {
    mockExecFileAsync.mockRejectedValue(new Error("bad"));

    const p = service.discover("/ws", ["./..."]);
    await vi.advanceTimersByTimeAsync(2_000);
    await vi.advanceTimersByTimeAsync(4_000);
    await p;

    expect(outputChannel.debug).toHaveBeenCalledWith(
      expect.stringContaining("attempt 1/3 failed, retrying"),
    );
    expect(outputChannel.debug).toHaveBeenCalledWith(
      expect.stringContaining("attempt 2/3 failed, retrying"),
    );
    expect(outputChannel.error).toHaveBeenCalledWith(
      expect.stringContaining("failed after 3 attempts"),
    );
    expect(outputChannel.error).toHaveBeenCalledTimes(1);
  });
});
