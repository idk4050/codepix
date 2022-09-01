package encoding

import "encoding/base64"

const encoding = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789AB"

var AlphaNumBase64 = base64.NewEncoding(encoding)
