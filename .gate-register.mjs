import { register } from "node:module";

register(new URL("./.gate-hook.mjs", import.meta.url));
