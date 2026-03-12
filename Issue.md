I tried running this with a valid API key and get no output:

```shell
go run translate.go -key=AIzaSy... 'Guten Tag!'
```

Through debugging it looks to me like the V2 API just doesn't work.

I added `fmt.Println(string(data))` after reading the body from the GET response and I get this error:

```json
"error": {
  "code": 403,
  "message": "Requests to this API translate method google.cloud.translate.v2.TranslateService.TranslateText are blocked.",
```

If I give it an invalid API key, `-key=foo`, I get:

```json
"error": {
  "code": 400,
  "message": "API key not valid. Please pass a valid API key.",
```

I've searched and searched and can find no mention of this problem, so maybe just me 🧐.

I've forked and have [a branch](https://github.com/zacharysyoung/robpike_translate/tree/translate_v3_svcaccount) that uses the V3 API and a Service Account credentials file to authenticate.
