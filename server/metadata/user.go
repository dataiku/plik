package metadata

import (
	"fmt"

	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"gorm.io/gorm"

	"github.com/root-gg/plik/server/common"
)

// CreateUser create a new user in DB
func (b *Backend) CreateUser(user *common.User) (err error) {
	return b.db.Create(user).Error
}

// UpdateUser update user info in DB
func (b *Backend) UpdateUser(user *common.User) (err error) {
	result := b.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(1) {
		return fmt.Errorf("no user updated")
	}

	return nil
}

// GetUser return a user from DB ( return nil and no error if not found )
func (b *Backend) GetUser(ID string) (user *common.User, err error) {
	user = &common.User{}
	err = b.db.Where(&common.User{ID: ID}).Take(user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return user, err
}

// GetUsers return all users
// provider is an optional filter
// admin is an optional filter ( nil = no filter, true = admins only, false = non-admins only )
func (b *Backend) GetUsers(provider string, admin *bool, withTokens bool, pagingQuery *common.PagingQuery) (users []*common.User, cursor *paginator.Cursor, err error) {
	if pagingQuery == nil {
		return nil, nil, fmt.Errorf("missing paging query")
	}

	p := pagingQuery.Paginator()
	p.SetKeys("CreatedAt", "ID")

	stmt := b.db.Model(&common.User{})

	if withTokens {
		stmt = stmt.Preload("Tokens")
	}

	if provider != "" {
		stmt = stmt.Where(&common.User{Provider: provider})
	}

	if admin != nil {
		// Use raw SQL instead of struct-based Where because GORM ignores zero-value
		// fields in structs, and false is the zero value for bool. Using the struct
		// pattern would silently skip the filter when querying for non-admin users.
		stmt = stmt.Where("is_admin = ?", *admin)
	}

	result, c, err := p.Paginate(stmt, &users)
	if err != nil {
		return nil, nil, err
	}
	if result.Error != nil {
		return nil, nil, result.Error
	}

	return users, &c, err
}

// SearchUsers returns users matching a LIKE query on id, login, name, and email.
// Results are always sorted by login and hard-capped at limit (max 20).
// provider and admin are optional filters, same as GetUsers.
func (b *Backend) SearchUsers(query string, provider string, admin *bool, limit int) (users []*common.User, err error) {
	if query == "" {
		return nil, fmt.Errorf("missing search query")
	}
	if limit <= 0 || limit > 20 {
		limit = 20
	}

	pattern := "%" + query + "%"
	stmt := b.db.Model(&common.User{}).
		Where("id LIKE ? OR login LIKE ? OR name LIKE ? OR email LIKE ?", pattern, pattern, pattern, pattern)

	if provider != "" {
		stmt = stmt.Where(&common.User{Provider: provider})
	}

	if admin != nil {
		stmt = stmt.Where("is_admin = ?", *admin)
	}

	stmt = stmt.Order("login ASC").Limit(limit)

	err = stmt.Find(&users).Error
	if err != nil {
		return nil, err
	}

	if users == nil {
		users = []*common.User{}
	}

	return users, nil
}

// ForEachUserUploads execute f for all upload matching the user and token filters
func (b *Backend) ForEachUserUploads(userID string, tokenStr string, f func(upload *common.Upload) error) (err error) {
	stmt := b.db.Model(&common.Upload{}).Where(&common.Upload{User: userID, Token: tokenStr})

	rows, err := stmt.Rows()
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		upload := &common.Upload{}
		err = b.db.ScanRows(rows, upload)
		if err != nil {
			return err
		}
		err = f(upload)
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveUserUploads deletes all uploads matching the user and token filters
func (b *Backend) RemoveUserUploads(userID string, tokenStr string) (removed int, err error) {
	deleted := 0
	var errors []error
	f := func(upload *common.Upload) (err error) {
		err = b.RemoveUpload(upload.ID)
		if err != nil {
			b.log.Warningf("unable to remove upload %s : %s", upload.ID, err)
			errors = append(errors, err)
			return nil
		}
		deleted++
		return nil
	}

	err = b.ForEachUserUploads(userID, tokenStr, f)
	if err != nil {
		return deleted, err
	}
	if len(errors) > 0 {
		return deleted, fmt.Errorf("unable to delete all user uploads")
	}

	return deleted, nil
}

// DeleteUser delete a user from the DB
func (b *Backend) DeleteUser(userID string) (deleted bool, err error) {
	_, err = b.RemoveUserUploads(userID, "")
	if err != nil {
		return false, err
	}

	err = b.db.Transaction(func(tx *gorm.DB) (err error) {
		// Delete user tokens
		err = tx.Where(&common.Token{UserID: userID}).Delete(&common.Token{}).Error
		if err != nil {
			return fmt.Errorf("unable to delete tokens metadata : %s", err)
		}

		// Delete user
		result := tx.Where(&common.User{ID: userID}).Delete(common.User{})
		if result.Error != nil {
			return fmt.Errorf("unable to delete user metadata : %s", result.Error)
		}

		if result.RowsAffected > 0 {
			deleted = true
		}

		return nil
	})

	return deleted, err
}

// CountUsers count the number of user in the DB
func (b *Backend) CountUsers() (count int, err error) {
	var c int64 // Gorm V2 needs int64 for counts
	err = b.db.Model(&common.User{}).Count(&c).Error
	if err != nil {
		return -1, err
	}

	return int(c), nil
}

// ForEachUsers execute f for every user in the database
func (b *Backend) ForEachUsers(f func(user *common.User) error) (err error) {
	rows, err := b.db.Model(&common.User{}).Rows()
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		user := &common.User{}
		err = b.db.ScanRows(rows, user)
		if err != nil {
			return err
		}
		err = f(user)
		if err != nil {
			return err
		}
	}

	return nil
}
