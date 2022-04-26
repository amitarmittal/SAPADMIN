package dream

/*
var (
	OperatorHttpReqTimeout time.Duration = 5
)

func PlaceBet_Transfer(betDto sports.BetDto) (float64, error) {
	var userBalance float64 = 0
	userKey := betDto.OperatorId + "-" + betDto.UserId
	// 6.2.1. Get user balance
	user, err := database.GetB2BUser(userKey)
	if err != nil {
		// 6.2.1.1. Get user balance failed
		log.Println("SportsBet: get user balance failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	userBalance = user.Balance
	// 6.2.2 Check for sufficient balance
	if betDto.BetReq.DebitAmount > userBalance {
		// 6.2.2.1. Get user balance failed
		log.Println("SportsBet: Insufficient Balance - ", user.Balance)
		//respDto.ErrorDescription = "Insufficient Balance"
		//respDto.Balance = user.Balance
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, fmt.Errorf("Insufficient Balance!")
	}
	// 6.2.3. Save in User Ledger
	userLedger := models.UserLedgerDto{}
	userLedger.UserKey = betDto.OperatorId + "-" + betDto.UserId
	userLedger.OperatorId = betDto.OperatorId
	userLedger.UserId = betDto.UserId
	userLedger.TransactionType = "Bet-Placement"
	userLedger.TransactionTime = time.Now().UnixNano() / int64(time.Millisecond)
	userLedger.ReferenceId = betDto.BetId
	userLedger.Amount = betDto.BetReq.DebitAmount * -1 // -ve value means, debited from user account
	err = database.InsertLedger(userLedger)
	if err != nil {
		// 6.2.3.1. inserting ledger document failed
		log.Println("SportsBet: insert ledger failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	// 6.2.4. Debit amount from user balance and save
	err = database.UpdateB2BUserBalance(userLedger.UserKey, userLedger.Amount)
	if err != nil {
		// 6.2.4.1. updating user balance failed
		log.Println("SportsBet: update user balance failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	// 6.2.5. set balance to resp object
	userBalance = userBalance - betDto.BetReq.DebitAmount
	return userBalance, nil
}

func PlaceBet_Seamless(betDto sports.BetDto, operatorDto operatordto.OperatorDTO, sessionDto sessdto.B2BSessionDto) (float64, error) {
	var userBalance float64 = 0
	// 6.2.2. Do operator's wallet call
	opRespDto, err := operator.WalletBet(betDto, sessionDto, operatorDto.Keys.PrivateKey, OperatorHttpReqTimeout)
	if err != nil {
		// 6.2.2.1. Wallet call failed
		log.Println("SportsBet: PlaceBet for Dream failed with error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	if opRespDto.Status != 0 {
		// 6.2.2.2. Wallet call returned error
		log.Println("SportsBet: Operator returned error for Wallet Bet. error - ", err.Error())
		//respDto.ErrorDescription = err.Error()
		//return c.Status(fiber.StatusOK).JSON(respDto)
		return userBalance, err
	}
	// 6.2.3. Set Balance
	userBalance = opRespDto.Balance
	return userBalance, nil
}
*/
