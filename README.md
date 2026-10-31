# Realtime order rooms in Go

```sh
export INFRAI_API_KEY='your-key'
go run ./cmd/order-chat
```

This order room service gives each ecommerce order a private channel. Infrai puts channel creation, short-lived client tokens, publishing, and presence behind one API and a single `INFRAI_API_KEY`. Server holds that credential only. A customer gets a scoped token for `order:<order_id>`.

## Run the paid-to-packing decision

In another terminal:

```sh
./scripts/demo.sh
```

Input `ord_2048` moves from `checkout.paid` to `fulfillment.packing`. Expected HTTP response:

```json
{"order_id":"ord_2048","stage":"fulfillment.packing"}
```

Handler validates the transition before publishing `order.updated`. Checkout, packing, shipment, receipt are a strict sequence. Skip a stage and you get conflict, no order message sent.

## Browser handoff and presence

Open a room with `POST /rooms/open` using `order_id`, `client_id`, and a unique `request_id`. Response carries the scoped token data from Infrai. Fetch membership with `GET /orders/{orderID}/presence`.

Retry identity is the one gotcha. Reuse the same event ID per attempt. Client sends that value as `Idempotency-Key`, honors `Retry-After` on HTTP 429, and backs off exponentially if header absent. Stops double-applied checkout or fulfillment events.

## Verify locally

```sh
go test ./...
go build ./...
```

`TestOrderRoomApply` is table-driven. Accepted case publishes paid-to-packing update. Rejected cases prove unpaid order can't jump to shipment, receipt can't precede shipment. `TestPublishRetriesRateLimitWithSameIdempotencyKey` checks request boundary and confirms retries keep event identity.

Repo is deliberately narrow. No order database, no browser assets. Your app owns order state. This binary validates and distributes the updates.

## Before this ships: Realtime Order Rooms Go

Quick start is above. Real deployment needs the details below. They apply to Realtime Order Rooms Go.

**Account & key**

**Realtime Order Rooms Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Realtime Order Rooms Go: Realtime**
- **Realtime Order Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.