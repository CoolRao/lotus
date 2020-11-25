package form

type IFrom interface {
	Valid() (error, bool)
}
