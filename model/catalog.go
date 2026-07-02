package model

import "time"

// SportType site'deki spor dalı seçeneğidir (ddlKiralikBransFiltre).
type SportType struct {
	ID        string    `json:"id"         gorm:"primaryKey;type:text"`
	Name      string    `json:"name"       gorm:"uniqueIndex;not null;type:text"` // görüntülenen ad
	SiteValue string    `json:"site_value" gorm:"not null;type:text"`              // dropdown value attr
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// Facility spor tipine bağlı tesis (ddlKiralikTesisFiltre).
type Facility struct {
	ID          string    `json:"id"           gorm:"primaryKey;type:text"`
	SportTypeID string    `json:"sport_type_id" gorm:"not null;index;type:text"`
	Name        string    `json:"name"         gorm:"not null;type:text"`
	SiteValue   string    `json:"site_value"   gorm:"not null;type:text"`
	CreatedAt   time.Time `json:"created_at"   gorm:"autoCreateTime"`

	SportType *SportType `json:"sport_type,omitempty" gorm:"foreignKey:SportTypeID"`
}

// Court tesise bağlı salon/kort (ddlKiralikSalonFiltre).
type Court struct {
	ID         string    `json:"id"          gorm:"primaryKey;type:text"`
	FacilityID string    `json:"facility_id" gorm:"not null;index;type:text"`
	Name       string    `json:"name"        gorm:"not null;type:text"`
	SiteValue  string    `json:"site_value"  gorm:"not null;type:text"`
	CreatedAt  time.Time `json:"created_at"  gorm:"autoCreateTime"`

	Facility *Facility `json:"facility,omitempty" gorm:"foreignKey:FacilityID"`
}
