package data

import (
	"context"
	"coursego/internal/validator"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Course struct {
	ID 			int64 				`json:"id"`	
	CreatedAt 	time.Time  			`json:"-"`
	Title 		string	 			`json:"title"`
	Year 		int32	 			`json:"year,omitempty"`
	Runtime     Runtime             `json:"-"`
	Subjects 	[]string 			`json:"subjects,omitempty"`
	Version 	int32 				`json:"version"`
}

func (m Course) MarshalJSON() ([]byte, error) {
	var runtime string

	if m.Runtime != 0 {
		runtime = fmt.Sprintf("%d mins", m.Runtime)
	}

	type CourseAlias Course
	aux := struct { 
		CourseAlias
		Runtime string `json:"runtime,omitempty"` 
	}{
		CourseAlias: CourseAlias(m),
		Runtime: runtime, 
	}

	return json.Marshal(aux)
}

func ValidateCourse(v *validator.Validator, course *Course) {
	v.Check(course.Title != "", "title", "must be provided")
	v.Check(len(course.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(course.Year != 0, "year", "must be provided")
	v.Check(course.Year >= 1888, "year", "must be greater than 1888")
	v.Check(course.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(course.Runtime != 0, "runtime", "must be provided") 
	v.Check(course.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(course.Subjects != nil, "subjects", "must be provided")
	v.Check(len(course.Subjects) >= 1, "subjects", "must contain at least 1 genre") 
	v.Check(len(course.Subjects) <= 5, "subjects", "must not contain more than 5 subjects") 
	v.Check(validator.Unique(course.Subjects), "subjects", "must not contain duplicate values")
}

type CourseModel struct { 
	DB *sql.DB
}

func (m CourseModel) Insert(course *Course) error { 
	query := `
		INSERT INTO courses (title, year, runtime, subjects) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []interface{}{course.Title, course.Year, course.Runtime, pq.Array(course.Subjects)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	return m.DB.QueryRowContext(ctx, query, args...).Scan(&course.ID, &course.CreatedAt, &course.Version)
}


func (m CourseModel) Get(id int64) (*Course, error) {
	if id < 1 {
		return nil, ErrRecordNotFound 
	}

	query := `
		SELECT id, created_at, title, year, runtime, subjects, version FROM courses
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var course Course

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&course.ID,
		&course.CreatedAt, 
		&course.Title,
		&course.Year, 
		&course.Runtime, 
		pq.Array(&course.Subjects), 
		&course.Version,
	)

	if err != nil {
		switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound 
			default:
				return nil, err 
		}
	}

	return &course, nil 
}

func (m CourseModel) Update(course *Course) error { 
	query := `
		UPDATE courses
		SET title = $1, year = $2, runtime = $3, subjects = $4, version = version + 1 WHERE id = $5 AND version = $6
		RETURNING version`

	args := []interface{} { 
		course.Title,
		course.Year, 
		course.Runtime, 
		pq.Array(course.Subjects), 
		course.ID,
		course.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&course.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict 
		default:
			return err 
		}
	}
	return nil
	
}

func (m CourseModel) Delete(id int64) error { 
	if id < 1 {
		return ErrRecordNotFound 
	}
	query := `
		DELETE FROM courses
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err 
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound 
	}

	return nil
}

func (m CourseModel) GetAll(title string, subjects []string, filters Filters) ([]*Course, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, runtime, subjects, version FROM courses
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') AND (subjects @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) 
	defer cancel()

	args := []interface{}{title, pq.Array(subjects), filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...) 
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	courses := []*Course{}

	for rows.Next() {
		var course Course
		err := rows.Scan(
			&totalRecords,
			&course.ID, 
			&course.CreatedAt, 
			&course.Title, 
			&course.Year, 
			&course.Runtime, 
			pq.Array(&course.Subjects), 
			&course.Version,
		)
		if err != nil {
			return nil, Metadata{}, err 
		}

		courses = append(courses, &course) 
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err 
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return courses, metadata, nil 
}