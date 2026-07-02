package catalogrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"appointment-scrapper/model"
)

type CatalogRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

// ─── SportType ────────────────────────────────────────────────────────────────

func (r *CatalogRepository) ListSportTypes(ctx context.Context) ([]*model.SportType, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, site_value, created_at FROM sport_types ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.SportType
	for rows.Next() {
		s := &model.SportType{}
		if err := rows.Scan(&s.ID, &s.Name, &s.SiteValue, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	if result == nil {
		result = []*model.SportType{}
	}
	return result, rows.Err()
}

// ─── Facility ─────────────────────────────────────────────────────────────────

func (r *CatalogRepository) ListFacilities(ctx context.Context, sportTypeID string) ([]*model.Facility, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, sport_type_id, name, site_value, created_at
		   FROM facilities WHERE sport_type_id=$1 ORDER BY name`,
		sportTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Facility
	for rows.Next() {
		f := &model.Facility{}
		if err := rows.Scan(&f.ID, &f.SportTypeID, &f.Name, &f.SiteValue, &f.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, f)
	}
	if result == nil {
		result = []*model.Facility{}
	}
	return result, rows.Err()
}

// ─── Court ───────────────────────────────────────────────────────────────────

func (r *CatalogRepository) ListCourts(ctx context.Context, facilityID string) ([]*model.Court, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, facility_id, name, site_value, created_at
		   FROM courts WHERE facility_id=$1 ORDER BY name`,
		facilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Court
	for rows.Next() {
		c := &model.Court{}
		if err := rows.Scan(&c.ID, &c.FacilityID, &c.Name, &c.SiteValue, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if result == nil {
		result = []*model.Court{}
	}
	return result, rows.Err()
}

// ─── Sync (upsert) ───────────────────────────────────────────────────────────

// UpsertCatalog katalog verisini transaction içinde tamamen yeniler.
func (r *CatalogRepository) UpsertCatalog(ctx context.Context, items []CatalogItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now()

	for _, item := range items {
		// sport_type
		stID := uuid.New().String()
		var existingStID string
		err := tx.QueryRow(ctx,
			`SELECT id FROM sport_types WHERE name=$1`, item.SportTypeName).Scan(&existingStID)
		if err != nil {
			// yok → insert
			if _, err := tx.Exec(ctx,
				`INSERT INTO sport_types(id, name, site_value, created_at) VALUES($1,$2,$3,$4)
				 ON CONFLICT(name) DO UPDATE SET site_value=EXCLUDED.site_value`,
				stID, item.SportTypeName, item.SportTypeSiteValue, now,
			); err != nil {
				return err
			}
			// tekrar oku
			_ = tx.QueryRow(ctx, `SELECT id FROM sport_types WHERE name=$1`, item.SportTypeName).Scan(&stID)
		} else {
			stID = existingStID
		}

		// facility
		fID := uuid.New().String()
		var existingFID string
		err = tx.QueryRow(ctx,
			`SELECT id FROM facilities WHERE sport_type_id=$1 AND name=$2`, stID, item.FacilityName).Scan(&existingFID)
		if err != nil {
			if _, err := tx.Exec(ctx,
				`INSERT INTO facilities(id, sport_type_id, name, site_value, created_at) VALUES($1,$2,$3,$4,$5)
				 ON CONFLICT DO NOTHING`,
				fID, stID, item.FacilityName, item.FacilitySiteValue, now,
			); err != nil {
				return err
			}
			_ = tx.QueryRow(ctx,
				`SELECT id FROM facilities WHERE sport_type_id=$1 AND name=$2`, stID, item.FacilityName).Scan(&fID)
		} else {
			fID = existingFID
		}

		// court
		if item.CourtName != "" {
			if _, err := tx.Exec(ctx,
				`INSERT INTO courts(id, facility_id, name, site_value, created_at) VALUES($1,$2,$3,$4,$5)
				 ON CONFLICT DO NOTHING`,
				uuid.New().String(), fID, item.CourtName, item.CourtSiteValue, now,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// CatalogItem katalog sync için en küçük birimi temsil eder.
type CatalogItem struct {
	SportTypeName      string
	SportTypeSiteValue string
	FacilityName       string
	FacilitySiteValue  string
	CourtName          string
	CourtSiteValue     string
}
