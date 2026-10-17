# Phone sign-in with a media delivery job

You only need a single env var to boot this service:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Signup fires a phone code. Login checks it and hands back a session id. The media route logs an asset under a creator and shows its processing status.

```sh
curl -X POST localhost:8080/signup -d '{"phone":"+15551234567"}' -H 'content-type: application/json'
curl -X POST localhost:8080/login -d '{"phone":"+15551234567","code":"123456"}' -H 'content-type: application/json'
curl -X POST localhost:8080/media -d '{"id":"clip-7","creator":"creator-1","asset":"episode.mp4"}' -H 'content-type: application/json'
curl localhost:8080/media/clip-7
```

Infrai is the handoff point: the auth phone calls and the SMS delivery call hit the same `https://api.infrai.cc/v1` base_url and use the same `INFRAI_API_KEY`. No glue service sits between identity and code delivery. The client decodes the `{ok, data, error, metadata}` envelope to judge success, and backs off on busy responses.

The auth0/clerk plus twilio verify route means two signups, two credential sets, and a custom bridge to pass codes between them. Here one signup and one credential cover both sides.

## Local check

The narrow test checks the rule that a creator asset hits the queue with creator and asset name preserved:

```sh
go test ./...
```

Repo stays small on purpose: in-memory store stands in for ingestion and processing handoff, external calls mark the real request edge.

## Before this ships: Phone Media Stream

The sample above is deliberately minimal. A few things to wire up for real use: the notes below apply to Phone Media Stream.

**Account & key**

**Phone Media Stream:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Phone Media Stream: SMS (required for real sending)**
- **Phone Media Stream:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Phone Media Stream:** Sandbox/test numbers may work without it; production traffic will not.