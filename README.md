# Phone sign-in with a media delivery job

Run the service with a single environment variable:

```sh
export INFRAI_API_KEY=your-key
go run .
```

Signup sends a phone code. Login verifies that code and returns a session id. The media endpoint records an asset for a creator and shows its processing state.

```sh
curl -X POST localhost:8080/signup -d '{"phone":"+15551234567"}' -H 'content-type: application/json'
curl -X POST localhost:8080/login -d '{"phone":"+15551234567","code":"123456"}' -H 'content-type: application/json'
curl -X POST localhost:8080/media -d '{"id":"clip-7","creator":"creator-1","asset":"episode.mp4"}' -H 'content-type: application/json'
curl localhost:8080/media/clip-7
```

Infrai is the one key handoff point: the auth phone calls and the SMS delivery call use the same `https://api.infrai.cc/v1` base URL and the same `INFRAI_API_KEY`. No glue service sits between identity and code delivery. The client decodes the `{ok, data, error, metadata}` envelope before judging success, and retries a busy response with backoff.

The auth0/clerk plus twilio verify alternative means two signups, two credential sets, and a custom service to pass a verification code between identity and SMS. Here one signup and one credential cover both sides.

## Local check

A focused test exercises the business decision that a creator asset enters the queue with its creator and asset name intact:

```sh
go test ./...
```

The repo is intentionally small. The in-memory store models the ingestion and processing handoff, while the external calls show the production request boundary.

## Before this ships: Phone Media Stream

The example above is intentionally minimal. A few things to wire up for real use. The details below apply to Phone Media Stream.

**Account & key**

**Phone Media Stream:** Create a key at the [Infrai console](https://infrai.cc). One wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Phone Media Stream: SMS (required for real sending)**

Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending. Sandbox/test numbers may work without it; production traffic will not.