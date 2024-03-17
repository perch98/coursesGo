ALTER TABLE courses ADD CONSTRAINT courses_runtime_check CHECK (runtime >= 0);
ALTER TABLE courses ADD CONSTRAINT courses_year_check CHECK (year BETWEEN 1888 AND date_part('year', now())); 
ALTER TABLE courses ADD CONSTRAINT subjects_length_check CHECK (array_length(subjects, 1) BETWEEN 1 AND 5);