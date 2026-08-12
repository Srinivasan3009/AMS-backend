-- ============================================================
-- Academic Management System — Schema + Seed Data
-- Safe to run on a fresh Postgres database. All statements are
-- idempotent (IF NOT EXISTS / ON CONFLICT DO NOTHING) so this can
-- also be re-run safely against a DB that already has this schema.
--
-- Test login credentials (all seeded faculty/student accounts):
--   Password: test1234
--   Faculty:  identifier = faculty_id   (e.g. FCS01)
--   Student:  identifier = register_no  (e.g. 2025CS01)
-- ============================================================


-- ============================================================
-- SECTION 1: SCHEMA
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    role VARCHAR(10) NOT NULL CHECK (role IN ('admin', 'faculty', 'student')),
    register_no VARCHAR(50) UNIQUE,
    faculty_id VARCHAR(50) UNIQUE,
    username VARCHAR(50) UNIQUE,
    name VARCHAR(100) NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS faculty_details (
    id SERIAL PRIMARY KEY,
    faculty_id VARCHAR(50) UNIQUE NOT NULL REFERENCES users(faculty_id),
    name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('Male', 'Female', 'Other')),
    designation VARCHAR(30) NOT NULL CHECK (designation IN ('Assistant Professor', 'Associate Professor', 'Professor')),
    department VARCHAR(10) NOT NULL DEFAULT 'CSE' CHECK (department IN ('CSE', 'IT', 'MECH', 'CIVIL')),
    mobile_number VARCHAR(15) NOT NULL,
    email VARCHAR(100) NOT NULL,
    address_1 VARCHAR(200),
    address_2 VARCHAR(200),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    date_of_retirement DATE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS student_details (
    id SERIAL PRIMARY KEY,
    register_no VARCHAR(50) UNIQUE NOT NULL REFERENCES users(register_no),
    name VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('Male', 'Female', 'Other')),
    father_name VARCHAR(100),
    mother_name VARCHAR(100),
    degree VARCHAR(10) NOT NULL DEFAULT 'B.E' CHECK (degree IN ('B.E', 'B.Tech', 'B.Sc', 'M.E', 'M.Tech')),
    department VARCHAR(10) NOT NULL CHECK (department IN ('CSE', 'IT', 'MECH', 'CIVIL')),
    batch VARCHAR(20) NOT NULL,
    joining_year VARCHAR(20) NOT NULL,
    mobile_number VARCHAR(15) NOT NULL,
    email VARCHAR(100) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS courses (
    id SERIAL PRIMARY KEY,
    course_no VARCHAR(20) UNIQUE NOT NULL,
    course_name VARCHAR(150) NOT NULL,
    department VARCHAR(10) NOT NULL CHECK (department IN ('CSE', 'IT', 'MECH', 'CIVIL')),
    semester INT NOT NULL CHECK (semester BETWEEN 1 AND 8),
    batch VARCHAR(20) NOT NULL,
    course_type VARCHAR(20) NOT NULL CHECK (course_type IN ('Theory', 'Lab', 'Theory+Lab')),
    course_category VARCHAR(10) NOT NULL CHECK (course_category IN ('Core', 'Elective')),
    lecture_hours INT NOT NULL DEFAULT 0,
    tutorial_hours INT NOT NULL DEFAULT 0,
    practical_hours INT NOT NULL DEFAULT 0,
    tcp INT NOT NULL,
    credit INT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS faculty_course_assignments (
    id SERIAL PRIMARY KEY,
    course_no VARCHAR(20) NOT NULL REFERENCES courses(course_no),
    faculty_id VARCHAR(50) NOT NULL REFERENCES faculty_details(faculty_id),
    term VARCHAR(10) NOT NULL,
    department VARCHAR(10) NOT NULL,
    semester INT NOT NULL,
    batch VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (course_no, term)
);

CREATE TABLE IF NOT EXISTS course_registrations (
    id SERIAL PRIMARY KEY,
    register_no VARCHAR(50) NOT NULL REFERENCES student_details(register_no),
    course_no VARCHAR(20) NOT NULL REFERENCES courses(course_no),
    term VARCHAR(10) NOT NULL,
    registered_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (register_no, course_no, term)
);

CREATE TABLE IF NOT EXISTS grades (
    id SERIAL PRIMARY KEY,
    register_no VARCHAR(50) NOT NULL REFERENCES student_details(register_no),
    course_no VARCHAR(20) NOT NULL REFERENCES courses(course_no),
    term VARCHAR(10) NOT NULL,
    grade VARCHAR(5) NOT NULL CHECK (grade IN ('O','A+','A','B+','B','C','U','RA','SA','W')),
    graded_by VARCHAR(50) NOT NULL REFERENCES faculty_details(faculty_id),
    graded_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (register_no, course_no, term)
);

CREATE TABLE IF NOT EXISTS registration_windows (
    id SERIAL PRIMARY KEY,
    feature_key VARCHAR(50) UNIQUE NOT NULL DEFAULT 'course_registration',
    start_datetime TIMESTAMP NOT NULL,
    end_datetime TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


-- ============================================================
-- SECTION 2: ADMIN ACCOUNT
-- ============================================================

INSERT INTO users (role, username, name, password_hash) VALUES
('admin', 'admin', 'Admin User', '$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6')
ON CONFLICT (username) DO NOTHING;
-- login: admin / test1234


-- ============================================================
-- SECTION 3: COURSES — CSE (Semesters I-IV, batch 2025-2029)
-- ============================================================

INSERT INTO courses (course_no, course_name, department, semester, batch, course_type, course_category, lecture_hours, tutorial_hours, practical_hours, tcp, credit, active) VALUES
('MA25C01','Applied Calculus','CSE',1,'2025-2029','Theory','Core',3,1,0,4,4,true),
('EN25C01','English Essentials - I','CSE',1,'2025-2029','Theory','Core',2,0,0,2,2,true),
('UC25H01','Heritage of Tamils','CSE',1,'2025-2029','Theory','Core',1,0,0,1,1,true),
('PH25C01','Applied Physics - I','CSE',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('CY25C01','Applied Chemistry - I','CSE',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('CS25C01','Computer Programming: C','CSE',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('CS25C03','Essentials of Computing','CSE',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('ME25C04','Makerspace','CSE',1,'2025-2029','Lab','Core',0,0,4,4,2,true),
('MA25C02','Linear Algebra','CSE',2,'2025-2029','Theory','Core',3,1,0,4,4,true),
('EE25C01','Basic Electrical and Electronics Engineering','CSE',2,'2025-2029','Theory','Core',3,0,0,3,3,true),
('CS25C06','Digital Principles and Computer Organization','CSE',2,'2025-2029','Theory','Core',3,1,0,4,4,true),
('UC25H02','Tamils and Technology','CSE',2,'2025-2029','Theory','Core',1,0,0,1,1,true),
('PH25C03','Applied Physics (CSIE) - II','CSE',2,'2025-2029','Theory','Core',2,1,0,3,3,true),
('CS25C07','Object Oriented Programming','CSE',2,'2025-2029','Theory+Lab','Core',3,0,4,7,5,true),
('EN25C02','English Essentials - II','CSE',2,'2025-2029','Theory+Lab','Core',1,0,2,3,2,true),
('ME25C05','Re-Engineering for Innovation','CSE',2,'2025-2029','Lab','Core',0,0,4,4,2,true),
('MA25C14','Discrete Mathematics','CSE',3,'2025-2029','Theory','Core',3,1,0,4,4,true),
('CS25C11','Operating Systems','CSE',3,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('CS25C10','Object Oriented Software Engineering','CSE',3,'2025-2029','Theory','Core',3,0,0,3,3,true),
('CS25C08','Data Structures','CSE',3,'2025-2029','Theory+Lab','Core',3,0,4,7,5,true),
('CS25C09','Java Programming','CSE',3,'2025-2029','Theory+Lab','Core',3,0,4,7,5,true),
('EN25C03','English Communication Skills Laboratory - I','CSE',3,'2025-2029','Lab','Core',0,0,2,2,1,true),
('MA25C13','Probability and Statistics','CSE',4,'2025-2029','Theory','Core',3,1,0,4,4,true),
('CS25C12','Algorithms','CSE',4,'2025-2029','Theory','Core',3,0,0,3,3,true),
('CS25C14','Theory of Computation','CSE',4,'2025-2029','Theory','Core',3,1,0,4,4,true),
('CS25C15','Standards in Computer Science','CSE',4,'2025-2029','Theory','Core',1,0,0,1,1,true),
('AD25201','Python for Data Science','CSE',4,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('CS25C13','Database Management Systems','CSE',4,'2025-2029','Theory+Lab','Core',3,0,4,7,5,true),
('EN25C04','English Communication Skills Laboratory - II','CSE',4,'2025-2029','Lab','Core',0,0,2,2,1,true)
ON CONFLICT (course_no) DO NOTHING;


-- ============================================================
-- SECTION 4: COURSES — MECH (Semesters I-IV, batch 2025-2029)
-- Includes MECH-specific rows (suffixed 'M') for subjects shared in
-- name with CSE but requiring their own course_no since course_no
-- is globally unique.
-- ============================================================

INSERT INTO courses (course_no, course_name, department, semester, batch, course_type, course_category, lecture_hours, tutorial_hours, practical_hours, tcp, credit, active) VALUES
('ME25C03','Introduction to Mechanical Engineering','MECH',1,'2025-2029','Theory','Core',2,1,0,3,3,true),
('ME25C01','Engineering Drawing','MECH',1,'2025-2029','Theory+Lab','Core',2,0,4,6,4,true),
('CS25C02','Computer Programming: Python','MECH',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('ME25C02','Engineering Mechanics','MECH',2,'2025-2029','Theory','Core',3,1,0,4,4,true),
('PH25C05','Applied Physics (ME) - II','MECH',2,'2025-2029','Theory','Core',2,1,0,3,3,true),
('CY25C03','Applied Chemistry (ME) - II','MECH',2,'2025-2029','Theory','Core',2,0,0,2,2,true),
('MA25C09','Computational Differential Equations','MECH',3,'2025-2029','Theory','Core',3,1,0,4,4,true),
('ME25C07','Applied Engineering Mechanics','MECH',3,'2025-2029','Theory','Core',3,0,0,3,3,true),
('ME25301','Engineering Thermodynamics','MECH',3,'2025-2029','Theory','Core',4,0,0,4,4,true),
('CE25C11','Strength of Materials','MECH',3,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('ME25C08','Metallurgy and Materials Science','MECH',3,'2025-2029','Theory','Core',3,0,0,3,3,true),
('EC25C17','Embedded Systems','MECH',3,'2025-2029','Theory','Core',3,0,0,3,3,true),
('CS25C16','Applied Data Science','MECH',4,'2025-2029','Theory','Core',3,0,0,3,3,true),
('ME25C09','Kinematics and Dynamics of Machines','MECH',4,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('CE25C12','Fluid Mechanics and Machinery','MECH',4,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('ME25401','Thermal Engineering - I','MECH',4,'2025-2029','Theory','Core',3,0,0,3,3,true),
('ME25402','Manufacturing Processes - I','MECH',4,'2025-2029','Theory+Lab','Core',3,0,2,5,4,true),
('ME25C10','Standards in Mechanical Engineering','MECH',4,'2025-2029','Theory','Core',1,0,0,1,1,true),
('MA25C01M','Applied Calculus','MECH',1,'2025-2029','Theory','Core',3,1,0,4,4,true),
('PH25C01M','Applied Physics - I','MECH',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('CY25C01M','Applied Chemistry - I','MECH',1,'2025-2029','Theory+Lab','Core',2,0,2,4,3,true),
('UC25H01M','Heritage of Tamils','MECH',1,'2025-2029','Theory','Core',1,0,0,1,1,true),
('EN25C01M','English Essentials - I','MECH',1,'2025-2029','Theory','Core',2,0,0,2,2,true),
('ME25C04M','Makerspace','MECH',1,'2025-2029','Lab','Core',0,0,4,4,2,true),
('MA25C02M','Linear Algebra','MECH',2,'2025-2029','Theory','Core',3,1,0,4,4,true),
('EE25C01M','Basic Electrical and Electronics Engineering','MECH',2,'2025-2029','Theory','Core',3,0,0,3,3,true),
('UC25H02M','Tamils and Technology','MECH',2,'2025-2029','Theory','Core',1,0,0,1,1,true),
('ME25C05M','Re-Engineering for Innovation','MECH',2,'2025-2029','Lab','Core',0,0,4,4,2,true),
('EN25C02M','English Essentials - II','MECH',2,'2025-2029','Theory+Lab','Core',1,0,2,3,2,true),
('EN25C03M','English Communication Skills Laboratory - I','MECH',3,'2025-2029','Lab','Core',0,0,2,2,1,true),
('EN25C04M','English Communication Skills Laboratory - II','MECH',4,'2025-2029','Lab','Core',0,0,2,2,1,true)
ON CONFLICT (course_no) DO NOTHING;


-- ============================================================
-- SECTION 5: FACULTY (users + faculty_details) — 12 total: 6 CSE, 6 MECH
-- ============================================================

INSERT INTO users (role, faculty_id, name, password_hash) VALUES
('faculty','FCS01','Dr. Kavitha Raman','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FCS02','Prof. Arun Kumar','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FCS03','Dr. Meena Sundaram','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FCS04','Mr. Vignesh Babu','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FCS05','Dr. Priya Sharma','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FCS06','Prof. Suresh Nair','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME01','Dr. Ganesh Iyer','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME02','Prof. Lakshmi Narayanan','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME03','Dr. Ramesh Chandran','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME04','Mr. Karthik Subramaniam','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME05','Dr. Deepa Venkatesh','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('faculty','FME06','Prof. Anand Krishnan','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6')
ON CONFLICT (faculty_id) DO NOTHING;

INSERT INTO faculty_details (faculty_id, name, date_of_birth, gender, designation, department, mobile_number, email, address_1, address_2, active, date_of_retirement) VALUES
('FCS01','Dr. Kavitha Raman','1978-03-14','Female','Professor','CSE','9840012301','kavitha.raman@annauniv.edu','12 Anna Nagar','Chennai',TRUE,NULL),
('FCS02','Prof. Arun Kumar','1985-07-22','Male','Associate Professor','CSE','9840012302','arun.kumar@annauniv.edu','45 T Nagar','Chennai',TRUE,NULL),
('FCS03','Dr. Meena Sundaram','1980-11-05','Female','Professor','CSE','9840012303','meena.sundaram@annauniv.edu','7 Adyar','Chennai',TRUE,NULL),
('FCS04','Mr. Vignesh Babu','1990-02-18','Male','Assistant Professor','CSE','9840012304','vignesh.babu@annauniv.edu','23 Velachery','Chennai',TRUE,NULL),
('FCS05','Dr. Priya Sharma','1983-09-30','Female','Associate Professor','CSE','9840012305','priya.sharma@annauniv.edu','56 Mylapore','Chennai',TRUE,NULL),
('FCS06','Prof. Suresh Nair','1975-05-12','Male','Professor','CSE','9840012306','suresh.nair@annauniv.edu','89 Guindy','Chennai',FALSE,'2035-05-12'),
('FME01','Dr. Ganesh Iyer','1979-01-25','Male','Professor','MECH','9840012401','ganesh.iyer@annauniv.edu','15 Tambaram','Chennai',TRUE,NULL),
('FME02','Prof. Lakshmi Narayanan','1986-06-08','Female','Associate Professor','MECH','9840012402','lakshmi.n@annauniv.edu','33 Porur','Chennai',TRUE,NULL),
('FME03','Dr. Ramesh Chandran','1981-10-17','Male','Professor','MECH','9840012403','ramesh.chandran@annauniv.edu','67 Ambattur','Chennai',TRUE,NULL),
('FME04','Mr. Karthik Subramaniam','1991-04-03','Male','Assistant Professor','MECH','9840012404','karthik.s@annauniv.edu','21 Chromepet','Chennai',TRUE,NULL),
('FME05','Dr. Deepa Venkatesh','1984-12-20','Female','Associate Professor','MECH','9840012405','deepa.venkatesh@annauniv.edu','44 Perungudi','Chennai',TRUE,NULL),
('FME06','Prof. Anand Krishnan','1976-08-27','Male','Professor','MECH','9840012406','anand.krishnan@annauniv.edu','78 Anna Salai','Chennai',FALSE,'2036-08-27')
ON CONFLICT (faculty_id) DO NOTHING;


-- ============================================================
-- SECTION 6: STUDENTS (users + student_details) — 15 total: 8 CSE, 7 MECH
-- ============================================================

INSERT INTO users (role, register_no, name, password_hash) VALUES
('student','2025CS01','Aditya Krishnan','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS02','Divya Prakash','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS03','Rahul Sivakumar','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS04','Sneha Balaji','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS05','Karthik Raja','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS06','Nithya Sree','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS07','Vishal Anand','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025CS08','Pooja Ramesh','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME01','Mohammed Faizal','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME02','Swetha Murugan','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME03','Bharath Kumar','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME04','Anjali Devi','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME05','Dinesh Kanna','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME06','Preethi Lakshmi','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6'),
('student','2025ME07','Santosh Kumar','$2b$12$pz75E7IT/Qdfr4syTNbf8esZ7W8p0AavLAhnYH9UuOZJZAiqyjwx6')
ON CONFLICT (register_no) DO NOTHING;

INSERT INTO student_details (register_no, name, date_of_birth, gender, father_name, mother_name, degree, department, batch, joining_year, mobile_number, email, active) VALUES
('2025CS01','Aditya Krishnan','2007-04-12','Male','Krishnan Moorthy','Radha Krishnan','B.E','CSE','2025-2029','2025-07','9940011001','aditya.k@student.annauniv.edu',TRUE),
('2025CS02','Divya Prakash','2007-08-25','Female','Prakash Raj','Kavya Prakash','B.E','CSE','2025-2029','2025-07','9940011002','divya.p@student.annauniv.edu',TRUE),
('2025CS03','Rahul Sivakumar','2007-01-30','Male','Sivakumar Elango','Geetha Sivakumar','B.E','CSE','2025-2029','2025-07','9940011003','rahul.s@student.annauniv.edu',TRUE),
('2025CS04','Sneha Balaji','2007-06-14','Female','Balaji Ravi','Uma Balaji','B.E','CSE','2025-2029','2025-07','9940011004','sneha.b@student.annauniv.edu',TRUE),
('2025CS05','Karthik Raja','2007-03-08','Male','Raja Gopalan','Malathi Raja','B.E','CSE','2025-2029','2025-07','9940011005','karthik.r@student.annauniv.edu',TRUE),
('2025CS06','Nithya Sree','2007-09-19','Female','Sree Kumar','Vani Sree','B.E','CSE','2025-2029','2025-07','9940011006','nithya.s@student.annauniv.edu',TRUE),
('2025CS07','Vishal Anand','2007-11-02','Male','Anand Murthy','Latha Anand','B.E','CSE','2025-2029','2025-07','9940011007','vishal.a@student.annauniv.edu',TRUE),
('2025CS08','Pooja Ramesh','2007-05-27','Female','Ramesh Kannan','Shanthi Ramesh','B.E','CSE','2025-2029','2025-07','9940011008','pooja.r@student.annauniv.edu',TRUE),
('2025ME01','Mohammed Faizal','2007-02-11','Male','Abdul Rahman','Fathima Begum','B.E','MECH','2025-2029','2025-07','9940011101','faizal.m@student.annauniv.edu',TRUE),
('2025ME02','Swetha Murugan','2007-07-23','Female','Murugan Selvam','Kalaivani Murugan','B.E','MECH','2025-2029','2025-07','9940011102','swetha.m@student.annauniv.edu',TRUE),
('2025ME03','Bharath Kumar','2007-04-05','Male','Kumar Sundaresan','Devi Kumar','B.E','MECH','2025-2029','2025-07','9940011103','bharath.k@student.annauniv.edu',TRUE),
('2025ME04','Anjali Devi','2007-10-16','Female','Devi Prasad','Kamala Devi','B.E','MECH','2025-2029','2025-07','9940011104','anjali.d@student.annauniv.edu',TRUE),
('2025ME05','Dinesh Kanna','2007-01-09','Male','Kanna Rajendran','Selvi Kanna','B.E','MECH','2025-2029','2025-07','9940011105','dinesh.k@student.annauniv.edu',TRUE),
('2025ME06','Preethi Lakshmi','2007-08-31','Female','Lakshmi Narayan','Bhuvana Lakshmi','B.E','MECH','2025-2029','2025-07','9940011106','preethi.l@student.annauniv.edu',TRUE),
('2025ME07','Santosh Kumar','2007-06-22','Male','Kumar Velusamy','Jaya Kumar','B.E','MECH','2025-2029','2025-07','9940011107','santosh.k@student.annauniv.edu',TRUE)
ON CONFLICT (register_no) DO NOTHING;


-- ============================================================
-- SECTION 7: REGISTRATION WINDOW (open now, for testing course registration)
-- ============================================================

INSERT INTO registration_windows (feature_key, start_datetime, end_datetime)
VALUES ('course_registration', '2026-08-01 00:00:00', '2026-12-31 23:59:59')
ON CONFLICT (feature_key)
DO UPDATE SET start_datetime = EXCLUDED.start_datetime, end_datetime = EXCLUDED.end_datetime;


-- ============================================================
-- SECTION 8: FACULTY_COURSE_ASSIGNMENTS
-- Past terms: July 2025 (sem1), Jan 2026 (sem2). Current: July 2026 (sem3).
-- ============================================================

INSERT INTO faculty_course_assignments (course_no, faculty_id, term, department, semester, batch) VALUES
('CS25C01','FCS01','July 2025','CSE',1,'2025-2029'),
('MA25C01','FCS02','July 2025','CSE',1,'2025-2029'),
('CS25C03','FCS03','July 2025','CSE',1,'2025-2029'),
('CS25C07','FCS01','Jan 2026','CSE',2,'2025-2029'),
('MA25C02','FCS02','Jan 2026','CSE',2,'2025-2029'),
('CS25C06','FCS04','Jan 2026','CSE',2,'2025-2029'),
('CS25C08','FCS01','July 2026','CSE',3,'2025-2029'),
('CS25C09','FCS05','July 2026','CSE',3,'2025-2029'),
('CS25C11','FCS03','July 2026','CSE',3,'2025-2029'),
('ME25C03','FME01','July 2025','MECH',1,'2025-2029'),
('ME25C01','FME02','July 2025','MECH',1,'2025-2029'),
('CS25C02','FME04','July 2025','MECH',1,'2025-2029'),
('ME25C02','FME01','Jan 2026','MECH',2,'2025-2029'),
('PH25C05','FME03','Jan 2026','MECH',2,'2025-2029'),
('ME25301','FME01','July 2026','MECH',3,'2025-2029'),
('CE25C11','FME05','July 2026','MECH',3,'2025-2029')
ON CONFLICT (course_no, term) DO NOTHING;


-- ============================================================
-- SECTION 9: COURSE_REGISTRATIONS (past + current terms)
-- ============================================================

INSERT INTO course_registrations (register_no, course_no, term) VALUES
('2025CS01','CS25C01','July 2025'),
('2025CS01','MA25C01','July 2025'),
('2025CS02','CS25C01','July 2025'),
('2025CS02','CS25C03','July 2025'),
('2025CS01','CS25C07','Jan 2026'),
('2025CS01','MA25C02','Jan 2026'),
('2025CS02','CS25C06','Jan 2026'),
('2025CS01','CS25C08','July 2026'),
('2025CS01','CS25C09','July 2026'),
('2025CS02','CS25C11','July 2026'),
('2025CS03','CS25C08','July 2026'),
('2025ME01','ME25C03','July 2025'),
('2025ME01','ME25C01','July 2025'),
('2025ME02','CS25C02','July 2025'),
('2025ME01','ME25C02','Jan 2026'),
('2025ME02','PH25C05','Jan 2026'),
('2025ME01','ME25301','July 2026'),
('2025ME02','CE25C11','July 2026')
ON CONFLICT (register_no, course_no, term) DO NOTHING;


-- ============================================================
-- SECTION 10: GRADES (past terms only — current term left ungraded/in-progress)
-- ============================================================

INSERT INTO grades (register_no, course_no, term, grade, graded_by) VALUES
('2025CS01','CS25C01','July 2025','O','FCS01'),
('2025CS01','MA25C01','July 2025','A+','FCS02'),
('2025CS02','CS25C01','July 2025','A','FCS01'),
('2025CS02','CS25C03','July 2025','B+','FCS03'),
('2025CS01','CS25C07','Jan 2026','A+','FCS01'),
('2025CS01','MA25C02','Jan 2026','A','FCS02'),
('2025CS02','CS25C06','Jan 2026','B+','FCS04'),
('2025ME01','ME25C03','July 2025','A','FME01'),
('2025ME01','ME25C01','July 2025','B+','FME02'),
('2025ME02','CS25C02','July 2025','O','FME04'),
('2025ME01','ME25C02','Jan 2026','A+','FME01'),
('2025ME02','PH25C05','Jan 2026','A','FME03')
ON CONFLICT (register_no, course_no, term)
DO UPDATE SET grade = EXCLUDED.grade, graded_by = EXCLUDED.graded_by, updated_at = NOW();

-- ============================================================
-- END OF SEED SCRIPT
-- ============================================================
