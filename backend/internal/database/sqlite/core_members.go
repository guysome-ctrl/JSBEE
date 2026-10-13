package sqlite

import (
	"github.com/Ankumeah/JSBEE/backend/internal/database"

	"context"
	"database/sql"
	"errors"
	"uuid"
)

func (s *SqlxDBController) AddCoreMember(
	ctx context.Context,
	member database.CoreMember,
) error {
	query := s.db.Rebind(`
    INSERT INTO core_members
    (uuid, name, role, description, image_filename, sort_order)
    VALUES (?, ?, ?, ?, ?, ?);
  `)

	if _, err := s.db.ExecContext(
		ctx, query,
		member.UUID, member.Name, member.Role,
		member.Description, member.ImageFilename, member.SortOrder,
	); isUniqueViolation(err) {
		return database.ErrExistCoreMember
	} else if err != nil {
		return err
	}

	return nil
}

func (s *SqlxDBController) GetCoreMember(
	ctx context.Context,
	uuid uuid.UUID,
) (database.CoreMember, error) {
	query := s.db.Rebind(`
    SELECT uuid, name, role, description, image_filename, sort_order
    FROM core_members
    WHERE (uuid = ?);
  `)

	var member database.CoreMember
	err := s.db.QueryRowxContext(ctx, query, uuid).StructScan(&member)
	if errors.Is(err, sql.ErrNoRows) {
		return member, database.ErrInvalidCoreMember
	}

	return member, err
}

func (s *SqlxDBController) GetCoreMembers(
	ctx context.Context,
) ([]database.CoreMember, error) {
	query := `
    SELECT uuid, name, role, description, image_filename, sort_order
    FROM core_members
    ORDER BY sort_order, rowid;
  `

	var members []database.CoreMember = []database.CoreMember{}
	err := s.db.SelectContext(ctx, &members, query)

	return members, err
}

func (s *SqlxDBController) UpdateCoreMember(
	ctx context.Context,
	member database.CoreMember,
) error {
	query := s.db.Rebind(`
    UPDATE core_members
    SET name = ?, role = ?, description = ?
    WHERE (uuid = ?);
  `)

	res, err := s.db.ExecContext(
		ctx, query,
		member.Name, member.Role, member.Description, member.UUID,
	)
	if isUniqueViolation(err) {
		return database.ErrExistCoreMember
	} else if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected < 1 {
		return database.ErrInvalidCoreMember
	}

	return nil
}

func (s *SqlxDBController) UpdateCoreMemberPhoto(
	ctx context.Context,
	uuid uuid.UUID,
	imageFilename string,
) error {
	query := s.db.Rebind(`
    UPDATE core_members
    SET image_filename = ?
    WHERE (uuid = ?);
  `)

	res, err := s.db.ExecContext(ctx, query, imageFilename, uuid)
	if isUniqueViolation(err) {
		return database.ErrExistCoreMember
	} else if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	} else if affected < 1 {
		return database.ErrInvalidCoreMember
	}

	return nil
}

func (s *SqlxDBController) DeleteCoreMember(
	ctx context.Context,
	uuid uuid.UUID,
) (database.CoreMember, error) {
	member, err := s.GetCoreMember(ctx, uuid)
	if err != nil {
		return member, err
	}

	query := s.db.Rebind(`
    DELETE FROM core_members
    WHERE (uuid = ?);
  `)

	res, err := s.db.ExecContext(ctx, query, uuid)
	if err != nil {
		return database.CoreMember{}, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return database.CoreMember{}, err
	} else if affected < 1 {
		return database.CoreMember{}, database.ErrInvalidCoreMember
	}

	return member, nil
}
