
CREATE TABLE Books (
    ID TEXT PRIMARY KEY,
    Title TEXT NOT NULL, 
    Author TEXT NOT NULL,
    ISBN TEXT NOT NULL,
    "Year" INTEGER,
    "Status" TEXT NOT NULL DEFAULT 'Available' CHECK ("Status" IN ('Available', 'Issued'))
);

CREATE TABLE Users (
    ID TEXT PRIMARY KEY,
    "Name" TEXT NOT NULL,
    Email TEXT NOT NULL,
    "Password" TEXT NOT NULL,
    "Role" TEXT NOT NULL CHECK ("Role" IN ('admin', 'user')),
    RegistrationDate DATE NOT NULL  
);

CREATE TABLE Issue (
    ID TEXT PRIMARY KEY,
    BookID TEXT REFERENCES Books(ID) NOT NULL,
    ReaderID TEXT REFERENCES User(ID) NOT NULL,
    IssueDate DATE NOT NULL,  
    DueDate DATE NOT NULL,
    ReturnDate DATE 
);