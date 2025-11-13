package external

type DomainMessage struct {
	DomainId int `json:"domain_id"`
}

type PubSubPushRequest struct {
	Message struct {
		Data []byte `json:"data"` // ← ここを []byte にする
	} `json:"message"`
}
