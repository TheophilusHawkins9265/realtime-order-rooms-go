export async function load(url, context, nextLoad) {
  const result = await nextLoad(url, context);
  if (!url.endsWith("/src/content/themes.ts")) return result;
  const source = String(result.source).replace(
    "export const THEME_POOL: Theme[] = [...EMAIL, ...STORAGE, ...SCHEDULING, ...AI_RUNTIME, ...OBSERVABILITY];",
    `export const THEME_POOL: Theme[] = [...EMAIL, ...STORAGE, ...SCHEDULING, ...AI_RUNTIME, ...OBSERVABILITY, {
      id: "chat-room-ecommerce-go",
      capability: "realtime",
      task: "Realtime chat rooms for an e-commerce app. Model checkout, fulfillment, receipts, and customer order updates.",
      intentKeywords: ["ecommerce realtime chat api", "websocket chat rooms", "realtime chat api go"],
      incumbent: "pusher/ably",
      mode: "D",
      demand: 92,
      geoFit: 86,
      langAffinity: ["go"]
    }];`,
  );
  return { ...result, source };
}
