package checkout

import (
	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/features/vehicle"
	auth_service "carf-chorf.com/app/app/services/auth"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/app/services/user_service"
)

// CreatePurchase validates the selected package, links/creates the buyer, and
// creates a pending transaction ready for the payment page.
func CreatePurchase(current *auth_service.State, req PurchaseRequest) (*domaine.Transaction, error) {
	pkg, ok := Find(req.Credits)
	if !ok {
		return nil, ErrInvalidPackage
	}

	if current != nil {
		if err := req.validateAuthenticated(); err != nil {
			return nil, err
		}
	} else if err := req.validateGuest(); err != nil {
		return nil, err
	}

	cacheRecordID, err := resolveCacheRecordID(req.Token, current == nil)
	if err != nil {
		return nil, err
	}

	user, err := resolvePurchaseUser(current, req)
	if err != nil {
		return nil, err
	}

	return transaction_service.Create(user.ID, cacheRecordID, pkg.Credits, pkg.AmountCents)
}

// resolveCacheRecordID returns the optional vehicle record. Guest checkout
// requires a vehicle; authenticated checkout may buy credits without one.
func resolveCacheRecordID(token string, guest bool) (*uint, error) {
	if token == "" {
		if guest {
			return nil, ErrVehicleRequired
		}
		return nil, nil
	}

	record, err := vehicle.FindRecordByToken(token)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrVehicleNotFound
	}
	return &record.ID, nil
}

// resolvePurchaseUser returns the signed-in user, or creates/finds the guest
// user by email, and rejects inactive accounts.
func resolvePurchaseUser(current *auth_service.State, req PurchaseRequest) (*domaine.User, error) {
	var user *domaine.User
	var err error

	if current != nil {
		user, err = user_service.FindByID(current.UserID)
	} else {
		user, err = user_service.FindOrCreateByEmail(req.Email)
	}
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrInactiveUser
	}

	return user, nil
}
