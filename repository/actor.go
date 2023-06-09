package repository

import (
	"crud/entity"
	"crud/utils/auth"
	"errors"
	"gorm.io/gorm"
	"time"
)

type Actor struct {
	db *gorm.DB
}

func NewActor(dbCrud *gorm.DB) Actor {
	return Actor{
		db: dbCrud,
	}

}

type ActorInterfaceRepo interface {
	CreateActor(actor *entity.Actor) (*entity.Actor, error)
	GetActorById(id uint) (entity.Actor, error)
	GetActors(username string, page uint) ([]entity.Actor, error)
	UpdateActorById(actor *entity.Actor, id uint) (*entity.Actor, error)
	DeleteActorById(id uint) (entity.Actor, error)
	Login(actor *entity.Actor) (*entity.Actor, error)
	Register(actor *entity.Actor) (*entity.Actor, error)
	GetRegisterApproval() ([]entity.RegisterApproval, error)
	UpdateRegisterApprovalStatusById(reg *entity.RegisterApproval, id uint) (*entity.RegisterApproval, error)
	SetActivateAdminById(id uint) (entity.Actor, error)
	SetDeactivateAdminById(id uint) (entity.Actor, error)
}

func (repo Actor) CreateActor(actor *entity.Actor) (*entity.Actor, error) {
	err := repo.db.Model(&entity.Actor{}).Create(actor).Error
	return actor, err
}

func (repo Actor) GetActorById(id uint) (entity.Actor, error) {
	var actor entity.Actor
	repo.db.First(&actor, "id = ? ", id)
	return actor, nil
}

func (repo Actor) GetActors(username string, page uint) ([]entity.Actor, error) {
	var actor []entity.Actor

	query := repo.db
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	limit := 5
	if page > 0 {
		offset := (int(page) - 1) * limit
		query = query.Limit(limit).Offset(offset)
	}

	query.Find(&actor)

	return actor, nil
}

func (repo Actor) UpdateActorById(actor *entity.Actor, id uint) (*entity.Actor, error) {
	var err error
	res := repo.db.Model(&actor).Where("id = ?", id).Updates(actor)
	if res.RowsAffected == 0 {
		err = errors.New("id not found")
	}
	return actor, err
}

func (repo Actor) DeleteActorById(id uint) (entity.Actor, error) {
	var actor entity.Actor
	var err error
	res := repo.db.Where("id = ? ", id).Delete(&actor)
	if res.RowsAffected == 0 {
		err = errors.New("id not found")
	}
	return actor, err
}

func (repo Actor) Login(actor *entity.Actor) (*entity.Actor, error) {

	// Check if the user exists in the database
	var admin *entity.Actor
	if err := repo.db.Where("username = ?", actor.Username).First(&admin).Error; err != nil {
		err = errors.New("invalid username or password")
		return actor, err
	}

	//Verify the password
	if err := auth.VerifyLogin(admin.Password, actor, admin.Salt); err != nil {
		err = errors.New("invalid username or password")
		return actor, err
	}

	//Verify the password
	//if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(actor.Password+admin.Salt)); err != nil {
	//	err = errors.New("invalid username or password")
	//	return actor, err
	//}

	return admin, nil
}

func (repo Actor) Register(actor *entity.Actor) (*entity.Actor, error) {

	// Check if the user exists in the database
	var admin *entity.Actor
	if err := repo.db.Where("username = ?", actor.Username).First(&admin).Error; err == nil {
		err = errors.New("username already exists")
		return admin, err
	}

	tx := repo.db.Begin()

	if err := tx.Model(&entity.Actor{}).Create(&actor).Error; err != nil {
		tx.Rollback()
		return admin, err
	}

	// Initiate the register approval
	err := tx.Model(&entity.RegisterApproval{}).Create(&entity.RegisterApproval{
		AdminID: actor.ID,
		Status:  "pending",
	}).Error

	if err != nil {
		tx.Rollback()
		return admin, err
	}

	tx.Commit()

	return admin, err
}

func (repo Actor) GetRegisterApproval() ([]entity.RegisterApproval, error) {
	var registerApproval []entity.RegisterApproval
	repo.db.Find(&registerApproval)
	return registerApproval, nil
}

func (repo Actor) UpdateRegisterApprovalStatusById(reg *entity.RegisterApproval, id uint) (*entity.RegisterApproval, error) {
	var err error
	var registerApproval entity.RegisterApproval

	tx := repo.db.Begin()

	res := tx.Model(&reg).Where("id = ?", id).Updates(reg)
	if res.RowsAffected == 0 {
		err = errors.New("id not found or no changes made")
	}

	tx.First(&registerApproval, "id = ?", id)

	switch reg.Status {
	case "approved":
		err := tx.Model(&entity.Actor{}).Where("id = ?", registerApproval.AdminID).Updates(entity.Actor{
			IsVerified: "true",
			IsActive:   "true",
			UpdatedAt:  time.Now(),
		}).Error
		if err != nil {
			tx.Rollback()
			return reg, err
		}
	case "rejected":
		err := tx.Model(&entity.Actor{}).Where("id = ?", registerApproval.AdminID).Updates(entity.Actor{
			IsVerified: "false",
			IsActive:   "false",
			UpdatedAt:  time.Now(),
		}).Error
		if err != nil {
			tx.Rollback()
			return reg, err
		}
	}

	tx.Commit()

	return reg, err
}

func (repo Actor) SetActivateAdminById(id uint) (entity.Actor, error) {
	var err error
	var actor entity.Actor
	res := repo.db.Model(&actor).Where("id = ?", id).Updates(&entity.Actor{
		IsActive:  "true",
		UpdatedAt: time.Now(),
	})
	if res.RowsAffected == 0 {
		err = errors.New("id not found")
	}

	return actor, err
}

func (repo Actor) SetDeactivateAdminById(id uint) (entity.Actor, error) {
	var err error
	var actor entity.Actor
	res := repo.db.Model(&actor).Where("id = ?", id).Updates(&entity.Actor{
		IsActive:  "false",
		UpdatedAt: time.Now(),
	})
	if res.RowsAffected == 0 {
		err = errors.New("id not found")
	}

	return actor, err
}
