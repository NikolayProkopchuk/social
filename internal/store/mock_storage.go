package store

func NewMockStore() *Storage {
	return &Storage{
		Users: &MockUserStore{},
		Roles: &MockRoleStore{},
	}
}
