import * as path from "node:path";
import { readFile } from "node:fs/promises";

export async function readModulePath(dir: string): Promise<string | undefined> {
  try {
    const content = await readFile(path.join(dir, "go.mod"), "utf-8");
    const match = /^\s*module\s+(\S+)/m.exec(content);
    return match?.[1];
  } catch {
    return undefined;
  }
}
