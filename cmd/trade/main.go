package main

//
//import (
//	"cb_grok/internal/exchange"
//	"cb_grok/internal/strategy"
//	"cb_grok/internal/telegram"
//	"fmt"
//	"go.uber.org/fx"
//	"time"
//)
//
//func main() {
//	fx.New(
//		exchange.Module,
//		strategy.Module,
//		telegram.Module,
//		fx.Invoke(runTrading),
//	).Run()
//}
//
//func runTrading(
//	ex exchange.Exchange,
//	strat strategy.Strategy,
//	tg *telegram.TelegramService,
//) {
//	symbol := "BTC/USDT"
//	position := 0.0
//	amount := 0.001 // Размер позиции в BTC
//
//	for {
//		candles, err := ex.Load(symbol, "1h", 100)
//		if err != nil {
//			tg.SendMessage(fmt.Sprintf("Ошибка загрузки данных: %v", err))
//			time.Sleep(5 * time.Minute)
//			continue
//		}
//
//		params := strategy.StrategyParams{
//			ShortPeriod:      10,
//			LongPeriod:       20,
//			RSIPeriod:        14,
//			ATRPeriod:        14,
//			BuyRSIThreshold:  30,
//			SellRSIThreshold: 70,
//			EMAShortPeriod:   50,
//			EMALongPeriod:    200,
//			UseTrendFilter:   true,
//			UseRSIFilter:     true,
//			ADXPeriod:        14,
//			UseADXFilter:     false,
//			ADXThreshold:     25.0,
//			ATRThreshold:     0.0,
//		}
//
//		candles = strat.Apply(candles, params)
//		lastSignal := candles[len(candles)-1].Signal
//		price := candles[len(candles)-1].Close
//
//		balance, err := ex.FetchBalance()
//		if err != nil {
//			tg.SendMessage(fmt.Sprintf("Ошибка получения баланса: %v", err))
//		}
//
//		if position == 0 && lastSignal == 1 && balance["USDT"] >= amount*price {
//			err = ex.CreateOrder(symbol, "buy", amount, price)
//			if err != nil {
//				tg.SendMessage(fmt.Sprintf("Ошибка покупки: %v", err))
//			} else {
//				position = amount
//				tg.SendMessage(fmt.Sprintf("Куплено %f BTC по цене %.2f", amount, price))
//			}
//		} else if position > 0 && lastSignal == -1 {
//			err = ex.CreateOrder(symbol, "sell", amount, price)
//			if err != nil {
//				tg.SendMessage(fmt.Sprintf("Ошибка продажи: %v", err))
//			} else {
//				position = 0
//				tg.SendMessage(fmt.Sprintf("Продано %f BTC по цене %.2f", amount, price))
//			}
//		}
//
//		time.Sleep(1 * time.Hour)
//	}
//}
