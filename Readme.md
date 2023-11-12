# nntpclient

This package provides a simple NNTP client that conforms to [RFC 3977][rfc3977].
It also supports the following extensions:

- [AUTHINFO](https://datatracker.ietf.org/doc/html/rfc4643)
- [STARTTLS](https://datatracker.ietf.org/doc/html/rfc4642)
- Connecting with TLS enabled connections (roughly [RFC 8143][rfc8143])

## TODO

- [ ] handle response codes that result in the server hanging-up
- [ ] support pooling (maybe)
- [ ] [COMPRESS](https://datatracker.ietf.org/doc/html/rfc8054) (maybe)
- [ ] [RFC 3977][rfc3977] section 6.3 (article posting)
- [ ] [RFC 3977][rfc3977] section 8 (e.g. `OVER`)

[rfc3977]: https://datatracker.ietf.org/doc/html/rfc3977
[rfc8143]: https://datatracker.ietf.org/doc/html/rfc8143
