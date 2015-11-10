package smsc
import (
    "net/http"
    "net/url"
    "io/ioutil"
    "strings"
    "errors"
    "encoding/json"
    "fmt"
)

// "Объект" для работы с отправщиком SMS
type SMSC struct {

    // Общий конфиг
    prefix string
    user string
    password string
    sender string

    // Список сообщений
    messages []messageSms

    // Прочие опции
    doTransliteration bool
    charset string
}

// Получаем инстанс сервиса
func GetService(prefix string, user string, password string, sender string) *SMSC {
    service := SMSC{}
    service.prefix = prefix
    service.user = user
    service.password = password
    service.sender = sender
    service.charset = "utf-8"

    return &service
}

const (
    URL_SEND = "http://smsc.ru/sys/send.php"
    URL_BALANCE = "http://smsc.ru/sys/balance.php"
)

type messageSms struct {
    Phone string
    Text string
}

// Добавляем сообщение в очередь отправки
func (this *SMSC) AddSms(phone, text string) *SMSC {
    this.messages = append(this.messages, messageSms{
        Phone: phone,
        Text: text,
    })

    return this
}

// Отправка всех сообщений из очереди
// Возвращает следующие параметры:
// - ID сервиса smsc.ru
// - Число отправленных SMS
// - Ошибка, если есть
func (this *SMSC) Send() (int, int, error) {
    response, err := http.Get(this.getSendUrl())
    defer response.Body.Close()
    if err != nil {
        return 0, 0, errors.New("HTTP-Request error: " + err.Error())
    }

    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return 0, 0, errors.New("Error while reading HTTP-response: " + err.Error())
    }

    type responseError struct {
        ErrorMessage string `json:"error"`
        ErrorCode int `json:"error_code"`
    }
    type responseSuccess struct {
        Id int `json:"id"`
        Count int `json:"cnt"`
    }

    responseData := string(contents)

    //это ответ об ошибке
    if strings.Contains(responseData, "error_code") {
        errorResponse := responseError{}
        err := json.Unmarshal([]byte(responseData), &errorResponse)
        if err != nil {
            return 0, 0, errors.New("Error while parsing JSON-ERROR: " + err.Error())
        }

        return 0, 0, errors.New(errorResponse.ErrorMessage)
    }

    //все вроде бы успешно
    if strings.Contains(responseData, "\"cnt\"") {
        successResponse := responseSuccess{}
        err := json.Unmarshal([]byte(responseData), &successResponse)
        if err != nil {
            return 0, 0, errors.New("Error while parsing JSON-SUCCESS: " + err.Error())
        }

        if successResponse.Count != len(this.messages) {
            msg := fmt.Sprint(
                "Sent SMS count (%d) is not equal to passed (%d)",
                len(this.messages),
                successResponse.Count,
            )
            return 0, 0, errors.New(msg)
        }

        return successResponse.Id, successResponse.Count, nil
    }

    return 0, 0, errors.New("Can't detect result type")
}

// Очистка очереди отправки
func (this *SMSC) Clear() *SMSC {
    this.messages = this.messages[:0]
    return this
}

// Надо ли транслитерировать сообщение
func (this *SMSC) SetTransliteration(enable bool) *SMSC {
    this.doTransliteration = enable
    return this
}

// Строим урл запроса для отправки SMS
func (this *SMSC) getSendUrl() string {

    var lines []string
    for _, message := range this.messages {
        lines = append(lines, message.Phone + ":" + this.prefix + strings.Replace(message.Text, "\n", "\\n", -1))
    }

    trans := "0"
    if this.doTransliteration {
        trans = "1"
    }

    params := url.Values{}
    params.Set("login", this.user)
    params.Set("psw", this.password)
    params.Set("sender", this.sender)
    params.Set("list", strings.Join(lines, "\n"))
    params.Set("translit", trans)
    params.Set("charset", this.charset)
    params.Set("fmt", "3")

    return URL_SEND + "?" + params.Encode()
}

// Выставление кодировки (по умолчанию utf-8)
func (this *SMSC) SetCharset(charset string) *SMSC {
    this.charset = charset
    return this
}
