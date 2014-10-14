package main

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/paymentd/config"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

func setDefaults(ctx *service.Context) error {
	paymentDB := ctx.PaymentDB()
	systemPassword, err := getSystemPassword(paymentDB)
	if err != nil {
		log.Crit("error checking for system password", log15.Ctx{"err": err})
		return err
	}
	if systemPassword == nil {
		log.Warn("system password not set. will generate a new system password...")
		genPwd := config.DefaultPassword("")
		err = genPwd.Generate()
		if err != nil {
			log.Error("error generating system password", log15.Ctx{"err": err})
			return err
		}
		err = config.Set(paymentDB, genPwd.Entry())
		if err != nil {
			log.Crit("error setting default settings", log15.Ctx{"err": err})
			return err
		}
		log.Warn("new system password set. please change as soon as possible", log15.Ctx{"systemPassword": string(genPwd)})
	}
	if ctx.Keychain().KeyCount() == 0 {
		log.Warn("no authorization keys set. will generate a new one...")
		_, err = ctx.Keychain().GenerateKey()
		if err != nil {
			log.Crit("error generating a new authorization key", log15.Ctx{"err": err})
			return err
		}
		generated, err := ctx.Keychain().Key()
		if err != nil {
			log.Crit("error retrieving generated key", log15.Ctx{"err": err})
			return err
		}
		log.Warn("generated auth key. please make sure to dump the generated keys if you intend to keep using the keys.", log15.Ctx{
			"generatedAuthKey": generated,
		})
	}
	return nil
}

func getSystemPassword(db *sql.DB) (*config.Entry, error) {
	return config.EntryByNameDB(db, config.ConfigNameSystemPassword)
}
