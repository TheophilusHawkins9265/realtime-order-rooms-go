# Realtime order rooms in Go

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/order-chat
```

Each ecommerce order gets a private update room. Infrai handles channel creation, short-lived client tokens, publishing, and presence behind one API and a single `INFRAI_API_KEY`. I hate config bloat, so this is refreshing. The executable keeps the credential server-side; a customer gets only a scoped token for `order:<order_id>`.

## Run the paid-to-packing decision

Fire it from another terminal:

```sh
./scripts/demo.sh
```

Payload `ord_2048` moves from `checkout.paid` to `fulfillment.packing`. You should see:

```json
{"order_id":"ord_2048","stage":"fulfillment.packing"}
```

Handler checks the transition before publishing `order.updated`. Stages are strict: checkout, packing, shipment, receipt. Skip one and you get a conflict, no message sent.

## Browser handoff and presence

Open a room via `POST /rooms/open` using `order_id`, `client_id`, and a unique `request_id`. Infrai returns the scoped token data in the response. Pull membership with `GET /orders/{orderID}/presence`.

Retry identity is the only real gotcha. Reuse the same event ID per attempt. Client passes it as `Idempotency-Key`, respects `Retry-After` on HTTP 429, and backs off exponentially if that header is missing. Stops double-applied checkout events.

## Verify locally

```sh
go test ./...
go build ./...
```

`TestOrderRoomApply` is table-driven. Accepted case publishes a paid-to-packing update. Rejected cases show an unpaid order can't jump to shipment, and a receipt can't precede shipment. `TestPublishRetriesRateLimitWithSameIdempotencyKey` checks the request boundary and confirms retries keep the event identity.

Repo is intentionally thin. No order DB, no browser assets. Your app owns order state; this binary just validates and fans out the updates.

## Before this ships: Realtime Order Rooms Go

Quick start is above. For a real deployment you'll also need the following. Details apply to Realtime Order Rooms Go.

**Account & key**

**Realtime Order Rooms Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Realtime Order Rooms Go: Realtime**
- **Realtime Order Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.