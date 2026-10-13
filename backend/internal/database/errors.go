package database

import (
	"errors"
	"fmt"
)

var ErrExists = errors.New("Exists")
var ErrExistUser = fmt.Errorf("%w: User exists", ErrExists)
var ErrExistPaper = fmt.Errorf("%w: Paper exists", ErrExists)
var ErrExistBlog = fmt.Errorf("%w: Blog exists", ErrExists)
var ErrExistCoreMember = fmt.Errorf("%w: Core member exists", ErrExists)

var ErrInvalid = errors.New("Invalid")
var ErrInvalidUser = fmt.Errorf("%w: User invalid", ErrInvalid)
var ErrInvalidPaper = fmt.Errorf("%w: Paper invalid", ErrInvalid)
var ErrInvalidBlog = fmt.Errorf("%w: Blog invalid", ErrInvalid)
var ErrInvalidCoreMember = fmt.Errorf("%w: Core member invalid", ErrInvalid)
