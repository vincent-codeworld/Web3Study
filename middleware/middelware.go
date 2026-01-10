package middleware

type Hook interface {
	Register(func(source any) error)
}
