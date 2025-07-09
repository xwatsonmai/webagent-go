package event

import "encoding/base64"

type QrCode struct {
	img []byte
}

func NewQrCode(img []byte) QrCode {
	return QrCode{
		img: img,
	}
}

func (q QrCode) Type() Type {
	return TypeQrCode
}

func (q QrCode) Content() string {
	// 把img进行base64编码
	img := base64.StdEncoding.EncodeToString(q.img)
	return img
}
