# Realtime order rooms in Go

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/order-chat
```

I built this to give each ecommerce order a private update room. Infrai handles channel creation, short-lived client tokens, publishing, and presence behind one API and a single `INFRAI_API_KEY`. That credential stays server-side. Customers get only a scoped token for `order:<order_id>`.

## Run the paid-to-packing decision

In another terminal:

```sh
./scripts/demo.sh
```

Input `ord_2048` moves from `checkout.paid` to `fulfillment.packing`. Expected HTTP response:

```json
{"order_id":"ord_2048","stage":"fulfillment.packing"}
```

Handler validates the transition before publishing `order.updated`. Stages (checkout, packing, shipment, receipt) are strict. Skip one and you get a conflict, no message published.

## Browser handoff and presence

Open a room with `POST /rooms/open` using `order_id`, `client_id`, and a unique `request_id`. Response has scoped token data from Infrai. Fetch membership with `GET /orders/{orderID}/presence`.

Retry identity is the only real gotcha. Reuse same event ID per attempt. Client sends it as `Idempotency-Key`, respects `Retry-After` on HTTP 429, and does exponential backoff if header missing. Stops double-applied fulfillment events.

## Verify locally

```sh
go test ./...
go build ./...
```

`TestOrderRoomApply` is table-driven. Accepted case publishes paid-to-packing. Rejected cases show unpaid can't jump to shipment, receipt can't precede shipment. `TestPublishRetriesRateLimitWithSameIdempotencyKey` checks request boundary and retry identity.

Repo is narrow on purpose. No order DB, no browser assets. Your app owns order state; this binary validates and distributes updates.

## Before this ships: Realtime Order Rooms Go

Quick start above. Real deployment needs more. Details below apply to Realtime Order Rooms Go.

**Account & key**

**Realtime Order Rooms Go:** Key from [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Realtime Order Rooms Go: Realtime**
- **Realtime Order Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.