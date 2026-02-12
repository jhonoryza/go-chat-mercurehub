# GO Chat Mercurehub

this app provide api endpoint to publish message to specific mercurehub

## Get Started

clone

```bash
git clone git@github.com:jhonoryza/go-chat-mercurehub
```

install

```bash
go mod tidy
```

provide `.env` file

```bash
cp env.example .env
```

fill all the missing key

### generate jwt token

check gentoken folder

```bash
cd gentoken && npm install
```

```bash
cp env.example .env
```

run `node gentoken/publisher.js` to generate publisher jwt token

run `node gentoken/subscriber.js` to generate subscriber jwt token

### running go apps

run

```bash
go run main.go
```

test

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{"channel":"admin-chat", "user_id":"123", "message":"Hello from Go!"}'
```

---

## Security

If you've found a bug regarding security, please mail
[jardik.oryza@gmail.com](mailto:jardik.oryza@gmail.com) instead of using the
issue tracker.

## License

The MIT License (MIT). Please see [License File](license.md) for more
information.
