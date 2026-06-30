package campaign

type Repository interface {
	save(campaign *Campaign) error
}
