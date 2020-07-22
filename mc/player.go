package mc

type Player struct {
	UUID          [16]byte
	OfflineUUID   [16]byte
	Username      string
	Skin          string
	SkinSignature string
}
