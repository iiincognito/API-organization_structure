-- +goose Up
CREATE TABLE departments (
                             id          SERIAL PRIMARY KEY,
                             name        VARCHAR(200) NOT NULL,
                             parent_id   INTEGER,
                             created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE employees (
                           id             SERIAL PRIMARY KEY,
                           department_id  INTEGER NOT NULL,
                           full_name      VARCHAR(200) NOT NULL,
                           position       VARCHAR(200) NOT NULL,
                           hired_at       DATE,
                           created_at     TIMESTAMPTZ DEFAULT NOW()
);

-- Foreign keys
ALTER TABLE employees
    ADD CONSTRAINT fk_employees_department
        FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

ALTER TABLE departments
    ADD CONSTRAINT fk_departments_parent
        FOREIGN KEY (parent_id) REFERENCES departments(id) ON DELETE CASCADE;

-- УНИКАЛЬНОСТЬ: название внутри одного parent (включая корневые)
CREATE UNIQUE INDEX idx_departments_name_parent
    ON departments (COALESCE(parent_id, 0), name);

-- Дополнительные индексы для производительности
CREATE INDEX idx_employees_department_id ON employees(department_id);
CREATE INDEX idx_employees_created_at ON employees(created_at);
CREATE INDEX idx_employees_full_name ON employees(full_name);

-- CHECK constraints (защита от пустых строк)
ALTER TABLE departments ADD CONSTRAINT chk_dept_name CHECK (TRIM(name) <> '');
ALTER TABLE employees ADD CONSTRAINT chk_emp_fullname CHECK (TRIM(full_name) <> '');
ALTER TABLE employees ADD CONSTRAINT chk_emp_position CHECK (TRIM(position) <> '');

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;