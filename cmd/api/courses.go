package main

import (
	"coursego/internal/data"
	"coursego/internal/validator"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func (app *application) createCourseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title	 string		 `json:"title"`
		Year	 int32		 `json:"year"`
		Runtime  data.Runtime `json:"runtime"` 
		Subjects []string 	 `json:"subjects"`
	}
	

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	course := &data.Course { 
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Subjects: input.Subjects,
	}

	v := validator.New()

	if data.ValidateCourse(v, course); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Courses.Insert(course)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/courses/%d", course.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"course": course}, headers)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	fmt.Fprintf(w, "%+v\n", input) 
}

func (app *application) showCourseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	course, err := app.models.Courses.Get(id)
	if err != nil { 
		switch {
			case errors.Is(err, data.ErrRecordNotFound): 
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"course": course}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCourseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	course, err := app.models.Courses.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r) 
		default:
			app.serverErrorResponse(w, r, err) 
		}
		return
	}

	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Subjects []string `json:"subjects"`
	}

	err = app.readJSON(w, r, &input) 
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil { 
		course.Title = *input.Title
	}

	if input.Year != nil { 
		course.Year = *input.Year
	}

	if input.Runtime != nil {
		course.Runtime = *input.Runtime 
	}

	if input.Subjects != nil {
		course.Subjects = input.Subjects
	}
	

	v := validator.New()
	if data.ValidateCourse(v, course); !v.Valid() { 
		app.failedValidationResponse(w, r, v.Errors) 
		return
	}

	err = app.models.Courses.Update(course) 
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r) 
			default:
			app.serverErrorResponse(w, r, err) 
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(course.Version), 32) != r.Header.Get("X-Expected-Version") { 
			app.editConflictResponse(w, r)
			return
		} 
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"course": course}, nil) 
	if err != nil {
		app.serverErrorResponse(w, r, err) 
	}
}

func (app *application) deleteCourseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Courses.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r) 
		default:
			app.serverErrorResponse(w, r, err) 
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "course successfully deleted"}, nil) 
	if err != nil {
		app.serverErrorResponse(w, r, err) 
	}
}

func (app *application) listCoursesHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string
		Subjects []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Subjects = app.readCSV(qs, "subjects", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	courses, metadata, err := app.models.Courses.GetAll(input.Title, input.Subjects, input.Filters) 
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"courses": courses, "metadata": metadata}, nil) 
	if err != nil {
		app.serverErrorResponse(w, r, err) 
	}
}