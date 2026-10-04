import { cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";

const sitePath = process.argv[2];

if (!sitePath || !/^[a-z0-9-]+$/.test(sitePath)) {
  throw new Error("Expected a URL-safe site path argument.");
}

const source = path.resolve("out");
const destinationRoot = path.resolve(".cloudflare/assets");
const destination = path.join(destinationRoot, sitePath);

await rm(destinationRoot, { recursive: true, force: true });
await mkdir(destination, { recursive: true });
await cp(source, destination, { recursive: true });

// Wrangler only reads _redirects from the assets root (a nested copy would be
// served as a public file), and the site lives under /<sitePath>. Move the file
// up and prefix both sides of every site-relative rule.
const nestedRedirects = path.join(destination, "_redirects");
const redirects = await readFile(nestedRedirects, "utf8").catch((error) => {
  if (error.code === "ENOENT") return null;
  throw error;
});

if (redirects !== null) {
  const prefix = (target) => (target.startsWith("/") ? `/${sitePath}${target}` : target);
  const rules = redirects.split("\n").map((line) => {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) return trimmed;
    const [from, to, ...rest] = trimmed.split(/\s+/);
    if (!from?.startsWith("/") || !to) {
      throw new Error(`Unsupported _redirects rule: ${line}`);
    }
    return [prefix(from), prefix(to), ...rest].join(" ");
  });
  await writeFile(path.join(destinationRoot, "_redirects"), `${rules.join("\n").trim()}\n`);
  await rm(nestedRedirects);
}
