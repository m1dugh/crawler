package crawler

type Attachement map[string]string

type Attachements []Attachement

func NewAttachements() Attachements {
	return make(Attachements, 0)
}

type OnPageResultAdded func(
	body []byte,
	pageResults PageResult,
	domainResults DomainResultEntry,
) Attachement
