package data

import ( 
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found") 
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Courses CourseModel 
	Tokens TokenModel
	Permissions PermissionModel
	Users UserModel
}

func NewModels(db *sql.DB) Models {
	return Models {
		Courses: CourseModel{DB: db},
		Tokens: TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Users: UserModel{DB: db}, 
	} 
}