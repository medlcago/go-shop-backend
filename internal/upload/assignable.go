package upload

type URLAssignable interface {
	GetObjectKey() string
	SetURL(url string)
}

type PublicURLBuilder interface {
	PublicURL(objectKey string) string
}

func AssignPublicURLs[T URLAssignable](items []T, builder PublicURLBuilder) {
	for i := range items {
		AssignPublicURL(items[i], builder)
	}
}

func AssignPublicURL[T URLAssignable](item T, builder PublicURLBuilder) {
	item.SetURL(builder.PublicURL(item.GetObjectKey()))
}
