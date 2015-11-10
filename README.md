# smscgo
SMSC.RU sender GO package

Отправка SMS-сообщений через сервис smsc.ru

http://smsc.ru/api/http/

Пример использования:

```go
_, _, err := smsc.GetService("", "login", "md5_pass", "SENDER.ID").AddSms("79999999999", "Text").Send()
```

