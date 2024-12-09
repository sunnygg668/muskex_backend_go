package ops

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"muskex/gen/mproto/model"
	"muskex/rserver"
	"muskex/utils"
	"muskex/utils/signal"
	"strconv"
	"sync"
	"time"
)

func DeamonAll() {
	log.Println("DeamonAll start")
	defer log.Println("DeamonAll exit")
	shutdownChan, err := signal.Intercept()
	if err != nil {
		panic(err)
	}

	c := cron.New(cron.WithSeconds())
	_, err = c.AddFunc("2 0 0 * * *", CalcUserLevel)
	if err != nil {
		log.Println("Error adding func:", err)
		return
	}
	//1 超时的充值订单取消
	c.AddFunc("2 */5 * * * *", RechargeOrderTimeout)
	//2 理财钱包收益按小时分发
	c.AddFunc("2 0 * * * *", TimerManagementWalletIncomeHours)
	//3 定期计算会员的激活状态：*/30 * * * * php think action api/Task/calcIsActivation
	c.AddFunc("2 */30 * * * *", CalcExpireActivation)
	// 4.分红奖励分发：55 23 * * * php think action api/Task/bonusAward
	c.AddFunc("2 55 23 * * *", BonusAward)
	// 5.矿机产出：0 */1 * * * php think action api/Task/minerOrderIncome
	c.AddFunc("2 0 */1 * * *", TimerMinerOrderIncome)
	// 6.定期理财结束，收益发放和本金返还：0 */1 * * * php think action api/Task/managementOrderIncome
	c.AddFunc("2 0 */1 * * *", TimerManagementOrderIncome)

	//7.每日统计报表生成：0 */2 * * * php think action api/StatisticsTask/reportStatistics
	c.AddFunc("2 0 */2 * * *", ReportStatistics)
	//8.总统计报表生成：0 1 * * * php think action api/StatisticsTask/reportStatisticsTotal
	c.AddFunc("2 0 1 * * *", ReportStatisticsTotal)

	////9.团队报表的数据生成：0 */2 * * * php think action api/StatisticsTask/reportTeamStatistics
	c.AddFunc("2 0 */2 * * *", ReportTeamStatistics)
	////10.团队总的报表数据生成：30 1 * * * php think action api/StatisticsTask/reportTeamStatisticsTotal
	c.AddFunc("2 30 1 * * *", ReportTeamStatisticsTotal)

	//11 理财返佣按小时分发.：0 23 * * * php think action api/Task/managementOrderRebateIncomeDay
	c.AddFunc("2 0 23 * * *", ManagementOrderRebateIncomeDay)
	c.AddFunc("2 * * * * *", CalcUserLevel)
	c.Start()

	<-shutdownChan.ShutdownChannel()
	for {
		size := 0
		workingMap.Range(func(key, value interface{}) bool {
			size++
			log.Println(key, " is working")
			return true
		})
		if size == 0 {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

var workingMap = sync.Map{}

/*
  - 定期理财结束，收益发放和本金返还，每小时执行一次

php code:
public function managementOrderIncome()

	{
	    Db::startTrans();
	    try {
	        $orderList = ManagementOrder::where(['status' => '1'])->whereTime('expire_time', '<=', time())->order('id asc')->select();
	        foreach ($orderList as $order) {
	            $totalPrice = $order->total_price;
	            $totalIncome = $order->total_income;
	            Assets::updateCoinAssetsBalance($order->user_id, $order->income_coin_id, $totalIncome, 'management_income');
	            Assets::updateCoinAssetsBalance($order->user_id, $order->settlement_coin_id, $totalPrice, 'management_total_price_return');
	            $order->paid_income += $totalIncome;
	            $order->status = '2';
	            $order->save();
	        }
	        Db::commit();
	    } catch (\Exception $e) {
	        Db::rollback();
	        $this->error($e->getMessage());
	    }
	    $this->success('理财收益已分发');
	}
*/
func TimerManagementOrderIncome() {
	workingMap.Store("TimerManagementOrderIncome", true)
	defer workingMap.Delete("TimerManagementOrderIncome")
	log.Println("TimerManagementOrderIncome start")
	defer log.Println("TimerManagementOrderIncome finish")
	tx := utils.Orm.Begin()
	defer tx.Rollback()

	var orderList []model.ManagementOrder
	if err := tx.Where("status = ?", "1").Where("expire_time <= ?", time.Now()).Order("id asc").Find(&orderList).Error; err != nil {
		log.Println("Error fetching orders:", err)
		return
	}

	for _, order := range orderList {
		totalPrice := order.TotalPrice
		totalIncome := order.TotalIncome

		if err := rserver.UpdateCoinAssetsBalance(tx, order.UserId, order.IncomeCoinId, totalIncome, "management_income", 0, 0, ""); err != nil {
			log.Println("Error updating coin assets balance for management_income:", err)
			return
		}

		if err := rserver.UpdateCoinAssetsBalance(tx, order.UserId, order.SettlementCoinId, totalPrice, "management_total_price_return", 0, 0, ""); err != nil {
			log.Println("Error updating coin assets balance for management_total_price_return:", err)
			return
		}

		order.PaidIncome += totalIncome
		order.Status = "2"

		if err := tx.Save(&order).Error; err != nil {
			log.Println("Error saving order:", err)
			return
		}
	}

	tx.Commit()
	log.Println("Management order income distributed successfully")
}

/*
*
  - 矿机到期发放本金和收益，每小时执行一次

php code:

	public function minerOrderIncome()
	   {
	       Db::startTrans();
	       try {
	           $orderList = Order::where(['status' => '1'])->whereTime('expire_time', '<=', time())->order('id asc')->select();
	           foreach ($orderList as $order) {
	               $realPay = $order->real_pay;
	               $estimatedIncome = $order->estimated_income;
	               Assets::updateCoinAssetsBalance($order->user_id, $order->produce_coin_id, $estimatedIncome, 'miners_income');
	               Assets::updateCoinAssetsBalance($order->user_id, $order->settlement_coin_id, $realPay, 'miners_real_pay_return');
	               $order->gained_income += $estimatedIncome;
	               $order->pending_income -= $estimatedIncome;
	               $order->status = '2';
	               $order->save();
	           }
	           Db::commit();
	       } catch (\Exception $e) {
	           Db::rollback();
	           $this->error($e->getMessage());
	       }
	       $this->success('矿机已产出');
	   }
*/
func TimerMinerOrderIncome() {
	workingMap.Store("TimerMinerOrderIncome", true)
	defer workingMap.Delete("TimerMinerOrderIncome")
	log.Println("TimerMinerOrderIncome start")
	defer log.Println("TimerMinerOrderIncome finish")
	tx := utils.Orm.Begin()
	defer tx.Rollback()

	var orderList []model.MinersOrder
	if err := tx.Where("status = ?", "1").Where("expire_time <= ?", time.Now()).Order("id asc").Find(&orderList).Error; err != nil {
		log.Println("Error fetching orders:", err)
		return
	}
	for _, order := range orderList {
		realPay := order.RealPay
		estimatedIncome := order.EstimatedIncome

		if err := rserver.UpdateCoinAssetsBalance(tx, order.UserId, order.ProduceCoinId, estimatedIncome, "miners_income", 0, 0, ""); err != nil {
			log.Println("Error updating coin assets balance for miners_income:", err)
			return
		}

		if err := rserver.UpdateCoinAssetsBalance(tx, order.UserId, order.SettlementCoinId, realPay, "miners_real_pay_return", 0, 0, ""); err != nil {
			log.Println("Error updating coin assets balance for miners_real_pay_return:", err)
			return
		}
		order.GainedIncome += estimatedIncome
		order.PendingIncome -= estimatedIncome
		order.Status = "2"
		if err := tx.Save(&order).Error; err != nil {
			log.Println("Error saving order:", err)
			return
		}
	}
	tx.Commit()
	log.Println("Miner order income distributed successfully")
}

/*
	理财钱包收益脚本，每小时执行一次

php code:

		public function managementWalletIncomeHours()
	    {
	        $moneyHourIncomeRatio = get_sys_config('money_hour_income_ratio');
	        $moneyDayIncomeRatio = bcmul($moneyHourIncomeRatio / 100, 1, 4);
	        if ($moneyDayIncomeRatio <= 0) {
	            $this->error('理财收益比例为 0，无需发放');
	        }
	        $hourTime = date('Y-m-d H');
	        Db::startTrans();
	        try {

	            $userList = User::where('status', 1)->where('money', '>', 0)->select();
	            foreach ($userList as $user) {

	                $first_transfer_in_where =[
	                    'user_id' => $user->id,
	                    'type' => 'transfer_in_money',
	                ];
	                $first_transfer_in = ManagementChange::where($first_transfer_in_where)->order('id asc')->find();
	                if ($first_transfer_in && (time() - $first_transfer_in->create_time < 3600*8 )  ) {
	                    echo $user->id."首次存入的每笔的收益，从8小时后才开始计算收益\r\n";
	                    continue;
	                }

	                $exist_where =[
	                    'user_id' => $user->id,
	                    'type' => 'management_income',
	                    'remark'=>$hourTime
	                ];
	                $exists = ManagementChange::where($exist_where)->find();
	                if ($exists) {
	                    //当前小时的理财收益已分发过，请勿重复操作
	                    echo $user->id."当前小时的理财收益已分发过，请勿重复操作\r\n";
	                    continue;
	                }

	                $income = bcmul($user->money, $moneyDayIncomeRatio, 2);
	                if ($income <= 0) {
	                    echo $user->id."理财收益为 0，无需发放\r\n";
	                    continue;
	                }

	                User::updateManagement($user->id, $income, 'management_income',null,$hourTime);
	            }
	            Db::commit();
	        } catch (\Exception $e) {
	            Db::rollback();
	            $this->error($e->getMessage());
	        }
	        $this->success($hourTime.'理财钱包收益分发成功');
	    }
*/
func TimerManagementWalletIncomeHours() {
	workingMap.Store("TimerManagementWalletIncomeHours", true)
	defer workingMap.Delete("TimerManagementWalletIncomeHours")
	log.Println("TimerManagementWalletIncomeHours start")
	defer log.Println("TimerManagementWalletIncomeHours finish")
	cfg := rserver.LoadDbConfig().NameValues
	moneyHourIncomeRatio, _ := strconv.ParseFloat(cfg["money_hour_income_ratio"], 64)
	moneyHourIncomeRatio = moneyHourIncomeRatio / 100
	if moneyHourIncomeRatio <= 0 {
		log.Println("Financial income ratio is 0, no need to distribute")
		return
	}
	hourTime := time.Now().Format("2006-01-02 15")

	tx := utils.Orm.Begin()
	defer tx.Rollback()
	sql1 := `
-- Update all users' money
		UPDATE ba_user
		SET money = money + money* @rate
		WHERE status = 1 AND money > 0
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ba_user_management_change mc
		    WHERE mc.user_id = ba_user.id
		      AND mc.type = 'transfer_in_money'
		      AND (UNIX_TIMESTAMP() - mc.create_time) < 3600 * 8
		  )
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ba_user_management_change mc
		    WHERE mc.user_id = ba_user.id
		      AND mc.type = 'management_income'
		      AND mc.remark = @hourTime
		  );`
	sql2 := `
		INSERT INTO ba_user_management_change (user_id, amount, ` + "`" + `before` + "`" + `, after, type, from_user_id, remark, referrer_id)
		SELECT u.id, money* @rate, u.money, u.money + money* @rate, 'management_income', 0, @hourTime, 0
		FROM ba_user u
		WHERE u.status = 1 AND u.money > 0
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ba_user_management_change mc
		    WHERE mc.user_id = u.id
		      AND mc.type = 'transfer_in_money'
		      AND (UNIX_TIMESTAMP() - mc.create_time) < 3600 * 8
		  )
		  AND NOT EXISTS (
		    SELECT 1
		    FROM ba_user_management_change mc
		    WHERE mc.user_id = u.id
		      AND mc.type = 'management_income'
		      AND mc.remark = @hourTime
		  );
`
	err1 := tx.Exec(sql1, map[string]interface{}{"rate": moneyHourIncomeRatio, "hourTime": hourTime}).Error
	if err1 != nil {
		return
	}
	err2 := tx.Exec(sql2, map[string]interface{}{"rate": moneyHourIncomeRatio, "hourTime": hourTime}).Error
	if err2 != nil {
		return
	}
	tx.Commit()
	log.Println("理财钱包收益分发成功", err1, err2)
}

/*
php code:
public function managementOrderRebateIncomeDay()

	{
	    $dayTime = date('Y-m-d', time());
	    Db::startTrans();
	    try {
	        $orderList = ManagementOrder::where(['status' => '1'])->where('rebate_income', '>', 0)->whereTime('expire_time', '>', time())->order('id asc')->select();
	        foreach ($orderList as $order) {
	            $user = User::where('id', $order->user_id)->find();
	            if(!$user->refereeid){//发放给上级
	                continue;
	            }
	            $exist_where =[
	                'user_id' => $user->refereeid,
	                'type' => 'rebate_income',
	                'remark'=>$dayTime.'_'.$order->id
	            ];
	            $exists = ManagementChange::where($exist_where)->find();
	            if ($exists) {
	                //当天的理财收益已分发过，请勿重复操作
	                echo $user->refereeid."的订单".$order->id."当天的理财收益已分发过，请勿重复操作\r\n";
	                continue;
	            }

	            $income = bcdiv($order->rebate_income, $order->closed_days, 2);
	            if ($income <= 0) {
	                echo $user->id."理财返利收益为 0，无需发放\r\n";
	                continue;
	            }
	            User::updateManagement($user->refereeid, $income, 'rebate_income',null,$dayTime.'_'.$order->id);
	        }
	        Db::commit();
	    } catch (\Exception $e) {
	        Db::rollback();
	        $this->error($e->getMessage());
	    }
	    $this->success($dayTime.'理财返利收益已分发');
	}
*/
func ManagementOrderRebateIncomeDay() {
	workingMap.Store("ManagementOrderRebateIncomeDay", true)
	defer workingMap.Delete("ManagementOrderRebateIncomeDay")
	log.Println("ManagementOrderRebateIncomeDay start")
	defer log.Println("ManagementOrderRebateIncomeDay finish")
	dayTime := time.Now().Format("2006-01-02")
	tx := utils.Orm.Begin()
	defer tx.Rollback()

	var orderList []model.ManagementOrder
	if err := tx.Where("status = ?", "1").Where("rebate_income > ?", 0).Where("expire_time > ?", time.Now()).Order("id asc").Find(&orderList).Error; err != nil {
		log.Println("Error fetching orders:", err)
		return
	}

	for _, order := range orderList {
		var user model.User
		if err := tx.Where("id = ?", order.UserId).First(&user).Error; err != nil {
			log.Println("Error fetching user:", err)
			continue
		}
		if user.Refereeid == 0 {
			continue
		}

		var exists model.ManChange
		existWhere := map[string]interface{}{
			"user_id": user.Refereeid,
			"type":    "rebate_income",
			"remark":  dayTime + "_" + strconv.Itoa(int(order.Id)),
		}
		if err := tx.Where(existWhere).First(&exists).Error; err == nil {
			log.Printf("User %d's order %d rebate income already distributed\n", user.Refereeid, order.Id)
			continue
		}

		income := order.RebateIncome / float64(order.ClosedDays)
		if income <= 0 {
			log.Printf("User %d rebate income is 0, no need to distribute\n", user.Id)
			continue
		}
		refUser := model.User{}
		utils.Orm.First(&refUser, user.Refereeid)
		if err := rserver.UpdateCmBalance(tx, refUser, income, "rebate_income", 0, 0, dayTime+"_"+strconv.Itoa(int(order.Id))); err != nil {
			log.Println("Error updating management:", err)
			return
		}
	}
	tx.Commit()
	log.Println(dayTime + " rebate income distributed successfully")
}

/*

   // 计算会员的激活状态，每半小时执行一次
   public function calcIsActivation()
   {
       $userActivationCalcInterval = get_sys_config('user_activation_calc_interval');
       $beginTime = strtotime('-' . $userActivationCalcInterval . ' hour');
       $userList = User::where(['status' => 1])->order('id asc')->select();
       foreach ($userList as $user) {
           if ($user->activation_time) {
               $shouldCalcTime = strtotime('+' . $userActivationCalcInterval . ' hour', $user->activation_time);
               if ($shouldCalcTime > time()) {
                   continue;
               }
           }
           $contractCount = ContractOrder::where(['user_id' => $user->id])->whereBetweenTime('buy_time', $beginTime, time())->count();
           $managementCount = ManagementOrder::where(['user_id' => $user->id])->whereBetweenTime('create_time', $beginTime, time())->count();
           $minersCount = Order::where(['user_id' => $user->id])->whereBetweenTime('create_time', $beginTime, time())->count();
           if (!$contractCount && !$managementCount && !$minersCount && $user->is_activation == 1) {
               $user->save(['is_activation' => 0, 'activation_time' => null]);
               Queue::push('\app\custom\job\UserQueue@updateTeamLevel', ['user_id' => $user->id, 'num' => -1], 'user');
           }
       }
       $this->success(date('Y-m-d') . ' 定期计算会员的激活状态');
   }
*/

// CalcIsActivation
func CalcExpireActivation() {
	workingMap.Store("CalcExpireActivation", true)
	defer workingMap.Delete("CalcExpireActivation")
	log.Println("CalcExpireActivation start")
	defer log.Println("CalcExpireActivation finish")
	cfg := rserver.LoadDbConfig().NameValues
	userActivationCalcInterval, _ := strconv.Atoi(cfg["user_activation_calc_interval"])
	beginTime := time.Now().Add(-time.Duration(userActivationCalcInterval) * time.Hour).Unix()
	var userList []model.User
	if err := utils.Orm.Where(`
status = @status and is_activation=1 and activation_time<@beginTime and not exists(select 1
                 from dual
                 where exists(select id from ba_trade_contract_order where user_id = ba_user.Id and buy_time > @beginTime)
                    or exists(select id from ba_trade_management_order where user_id = ba_user.Id  and create_time > @beginTime)
                    or exists(select id from ba_miners_order where user_id = ba_user.Id  and create_time > @beginTime))`, map[string]interface{}{"status": "1", "beginTime": beginTime}).Order("id asc").Find(&userList).Error; err != nil {
		log.Println("Error fetching users:", err)
		return
	}
	for _, user := range userList {

		//contractCount := utils.Orm.Model(&model.ContractOrder{}).Where("user_id = ?", user.ID).Where("buy_time BETWEEN ? AND ?", beginTime, time.Now()).Count()
		//managementCount := utils.Orm.Model(&model.ManagementOrder{}).Where("user_id = ?", user.ID).Where("create_time BETWEEN ? AND ?", beginTime, time.Now()).Count()
		//minersCount := utils.Orm.Model(&model.Order{}).Where("user_id = ?", user.ID).Where("create_time BETWEEN ? AND ?", beginTime, time.Now()).Count()
		//if contractCount == 0 && managementCount == 0 && minersCount == 0 && user.IsActivation == 1 {
		//	user.IsActivation = 0
		//	user.ActivationTime = nil
		//	if err := utils.Orm.Save(&user).Error; err != nil {
		//		log.Println("Error updating user activation status:", err)
		//		continue
		//	}
		//	QueuePush("updateTeamLevel", map[string]interface{}{"user_id": user.ID, "num": -1}, "user")
		//}

		user.IsActivation = 0
		user.ActivationTime = 0
		if err := utils.Orm.Updates(&model.User{IsActivation: 0, ActivationTime: 0}).Error; err != nil {
			log.Println("Error updating user activation status:", err)
			continue
		}
		//todo  updateTeamLevel
		//QueuePush("updateTeamLevel", map[string]interface{}{"user_id": user.ID, "num": -1}, "user")
		r := rserver.Reward{}
		r.UpdateTeamLevelFroActive(user.Id, false)
	}

	log.Println(" user activation expire_status  calculated")
}

/*
//充值订单超时取消脚本，每5分钟执行一次
public function rechargeOrderTimeout()

	{
	    $rechargeTimeoutInterval = get_sys_config('recharge_timeout_interval');
	    $limitTime = strtotime('-' . $rechargeTimeoutInterval . ' minute');
	    FinancialRecharge::where(['status' => 0])->whereTime('create_time', '<=', $limitTime)->update(['status' => 2]);
	    $this->success('超时的充值订单已取消');
	}
*/
func RechargeOrderTimeout() {
	workingMap.Store("RechargeOrderTimeout", true)
	defer workingMap.Delete("RechargeOrderTimeout")
	log.Println("RechargeOrderTimeout start")
	defer log.Println("RechargeOrderTimeout finish")
	cfg := rserver.LoadDbConfig().NameValues
	rechargeTimeoutInterval, _ := strconv.Atoi(cfg["recharge_timeout_interval"])
	limitTime := time.Now().Add(-time.Duration(rechargeTimeoutInterval) * time.Minute)

	if err := utils.Orm.Model(&model.FinancialRecharge{}).Where("status = ?", 0).Where("create_time <= ?", limitTime).Update("status", 2).Error; err != nil {
		log.Println("Error updating recharge order status:", err)
		return
	}

	log.Println("Timeout recharge orders cancelled")
}

/*
//分红奖励，每天23:55执行
public function bonusAward()

	{
	    $levelMap = [];
	    $levelList = Level::where('is_open', 1)->where('bonus', '>', 0)->order('level asc')->select();
	    foreach ($levelList as $level) {
	        $levelMap[$level->level] = $level;
	    }
	    Db::startTrans();
	    try {
	        $userList = User::where('status', 1)->select();
	        foreach ($userList as $user) {
	            $exists = CoinChange::where('type', 'bonus_award')->where('user_id', $user->id)->whereDay('create_time')->find();
	            if ($exists) {
	                continue;
	            }
	            if (array_key_exists($user->level, $levelMap)) {
	                $level = $levelMap[$user->level];
	                $bonus = $level->bonus;
	                Assets::updateMainCoinAssetsBalance($user->id, $bonus, 'bonus_award');
	            }
	        }
	        Db::commit();
	    } catch (Exception $e) {
	        Db::rollback();
	        $this->error($e->getMessage());
	    }
	    $this->success('分红奖励分发成功');
	}
*/
func BonusAward() {
	workingMap.Store("BonusAward", true)
	defer workingMap.Delete("BonusAward")
	log.Println("BonusAward start")
	defer log.Println("BonusAward finish")
	var levelMap = make(map[uint32]model.UserLevel)
	var levelList []model.UserLevel
	if err := utils.Orm.Where("is_open = ?", 1).Where("bonus > ?", 0).Order("level asc").Find(&levelList).Error; err != nil {
		log.Println("Error fetching levels:", err)
		return
	}

	for _, level := range levelList {
		levelMap[level.Level] = level
	}

	tx := utils.Orm.Begin()
	defer tx.Rollback()

	var userList []model.User
	if err := tx.Where("status = ?", 1).Find(&userList).Error; err != nil {
		log.Println("Error fetching users:", err)
		return
	}

	for _, user := range userList {
		var exists model.UserCoinChange
		if err := tx.Where("type = ?", "bonus_award").Where("user_id = ?", user.Id).Where("DATE(from_unixtime(create_time)) = CURDATE()").First(&exists).Error; err == nil {
			continue
		}
		if level, ok := levelMap[uint32(user.Level)]; ok {
			bonus := float64(level.Bonus)
			if bonus <= 0 {
				continue
			}
			if err := rserver.UpdateCoinAssetsBalance(tx, user.Id, 1, bonus, "bonus_award", 0, 0, ""); err != nil {
				log.Println("Error updating main coin assets balance:", err)
				return
			}
		}
	}

	tx.Commit()
	log.Println("Bonus award distributed successfully")
}

func CalcUserLevel() {
	workingMap.Store("CalcUserLevel", true)
	defer workingMap.Delete("CalcUserLevel")
	log.Println("CalcUserLevel begin")
	defer log.Println("CalcUserLevel end")
	sql := `
WITH user_stats AS (
    SELECT
        pid AS user_id,
        COUNT(1) AS team_nums,
        SUM(CASE WHEN team_level = 1 THEN 1 ELSE 0 END) AS referee_nums
    FROM ba_team_user
    WHERE is_active = 1
    GROUP BY pid
)
UPDATE ba_user u
JOIN user_stats us ON u.id = us.user_id
SET
    level = (
        SELECT MAX(l.level)
        FROM ba_user_level l
        WHERE us.referee_nums >= l.referee_num
          AND us.team_nums >= l.team_num
    ),
    update_time = unix_timestamp()
WHERE EXISTS (
    SELECT 1
    FROM ba_user_level l
    WHERE us.referee_nums >= l.referee_num
      AND us.team_nums >= l.team_num
)
AND EXISTS (
    SELECT 1
    FROM ba_team_user
    WHERE u.id = ba_team_user.pid
      AND u.update_time > unix_timestamp() - 610
);
insert ba_team_user (id, pid, user_level,is_whitelist,is_active)
select ba_user.id, ba_user.refereeid pid, level,ba_user.is_whitelist,ba_user.is_activation
from ba_user
inner join ba_team_user on ba_user.id=ba_team_user.id
where ba_user.update_time > unix_timestamp() - 612
ON DUPLICATE KEY UPDATE user_level=ba_user.level,is_whitelist=ba_user.is_whitelist,is_active=ba_user.is_activation;`
	err := utils.Orm.Exec(sql).Error
	if err != nil {
		log.Println("Error updating user level:", err)
		return
	}
	log.Println("User level calculated successfully")
}

/*
public function reportStatistics()

	{
	    $date = date('Y-m-d');
	    $mainCoinId = get_sys_config('main_coin');
	    $activityTypes = ['check_in_reward', 'first_contract_amount_reached', 'month_invite_reached_give',
	        'week_invite_reached_give', 'today_invite_reached_give', 'month_contract_amount_reached',
	        'today_contract_amount_reached', 'today_contract_num_reached', 'team_num_reached_give',
	        'invite_num_reached_give', 'today_recharge_reached_give', 'first_recharge_reached_give',
	        'auth_give', 'u_recharge', 'invite_first_recharge', 'invite_register_reward', 'register_reward'];
	    $whitelistUserIds = User::where('is_whitelist', 1)->column('id');
	    $todayQuery = function (Query $query) use ($whitelistUserIds, $date) {
	        $query->whereDay('create_time', $date)->where('user_id', 'not in', $whitelistUserIds);
	    };
	    $updateTimeTodayQuery = function (Query $query) use ($whitelistUserIds, $date) {
	        $query->whereDay('update_time', $date)->where('user_id', 'not in', $whitelistUserIds);
	    };
	    $regNums = User::whereDay('create_time', $date)->where('id', 'not in', $whitelistUserIds)->count();
	    $minersIncome = CoinChange::where($todayQuery)->where('type', 'miners_income')->where('coin_id', $mainCoinId)->sum('amount');
	    $managementIncome = CoinChange::where($todayQuery)->where('type', 'management_income')->where('coin_id', $mainCoinId)->sum('amount');
	    $income = bcadd($minersIncome, $managementIncome, 2);
	    $minersConsumption = Order::where('settlement_coin_id', $mainCoinId)->where($todayQuery)->sum('total_price');
	    $managementConsumption = ManagementOrder::where('settlement_coin_id', $mainCoinId)->where($todayQuery)->sum('total_price');
	    $consumption = bcadd($minersConsumption, $managementConsumption, 2);
	    $rechargeCoin = Recharge::where($todayQuery)->sum('amount');
	    $rechargeMoney = FinancialRecharge::where('status', 1)->where($todayQuery)->sum('amount');
	    $managementBuy = $managementConsumption;
	    $minersBuy = $minersConsumption;
	    $withdraw = Withdraw::where($updateTimeTodayQuery)->where('status', 'in', '1,3')->where('type', 0)->sum('money');
	    $rebate = CommissionChange::where($todayQuery)->where('type', 'margin_reward')->sum('amount');
	    $activity = CoinChange::where($todayQuery)->where('type', 'in', $activityTypes)->sum('amount');
	    $withdrawFee = Withdraw::where($updateTimeTodayQuery)->where('status', 'in', '1,3')->sum('fee_coin');
	    $fee = $withdrawFee;
	    $payment = ContractOrder::whereDay('buy_time', $date)->where('user_id', 'not in', $whitelistUserIds)->where('payment_status', 1)->sum('income');
	    $payment = abs($payment);
	    $minersProduce = CoinChange::where($todayQuery)->where('type', 'miners_income')->where('coin_id', $mainCoinId)->sum('amount');
	    $data = [
	        'date' => $date,
	        'reg_nums' => $regNums,
	        'income' => $income,
	        'consumption' => $consumption,
	        'recharge_coin' => $rechargeCoin,
	        'recharge_money' => $rechargeMoney,
	        'management_buy' => $managementBuy,
	        'miners_buy' => $minersBuy,
	        'withdraw' => $withdraw,
	        'rebate' => $rebate,
	        'activity' => $activity,
	        'fee' => $fee,
	        'payment' => $payment,
	        'bonus' => 0,
	        'miners_produce' => $minersProduce,
	        'management_income' => $managementIncome,
	        'create_time' => time()
	    ];
	    $statistics = Statistics::where('date', $date)->find();
	    if ($statistics) {
	        $statistics->save($data);
	    } else {
	        Statistics::create($data);
	    }
	    $this->success($date . ' 的统计报表生成成功');
	}
*/
func ReportStatistics() {
	workingMap.Store("ReportStatistics", true)
	defer workingMap.Delete("ReportStatistics")
	log.Println("ReportStatistics begin")
	defer log.Println("ReportStatistics end")
	sql_stat := `
INSERT INTO ba_report_statistics (date, reg_nums, income, consumption, recharge_coin, recharge_money, management_buy, miners_buy, withdraw, rebate, activity, fee, payment, bonus, miners_produce, management_income, create_time)
WITH whitelist AS (
		SELECT id from ba_user where is_whitelist =0
	)
SELECT
		 from_unixtime(@date) date,
		(SELECT COUNT(*) FROM ba_user WHERE truncate(create_time,86400) = @date AND id  IN (SELECT id FROM whitelist)) AS reg_nums,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE truncate(create_time,86400) = @date AND type = 'miners_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) +
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE truncate(create_time,86400) = @date AND type = 'management_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS income,
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_miners_order WHERE truncate(create_time,86400) = @date AND settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) +
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_trade_management_order WHERE truncate(create_time,86400) = @date AND settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS consumption,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_coin_recharge WHERE truncate(create_time,86400) = @date AND user_id  IN (SELECT id FROM whitelist)) AS recharge_coin,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_financial_recharge WHERE status = 1 AND truncate(create_time,86400) = @date AND user_id  IN (SELECT id FROM whitelist)) AS recharge_money,
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_trade_management_order WHERE truncate(create_time,86400)= @date AND settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS management_buy,
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_miners_order WHERE truncate(create_time,86400)= @date AND settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS miners_buy,
		(SELECT COALESCE(SUM(money), 0) FROM ba_financial_withdraw WHERE truncate(update_time,86400) = @date AND status IN (1, 3) AND type = 0 AND user_id  IN (SELECT id FROM whitelist)) AS withdraw,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_commission_change WHERE truncate(create_time,86400) = @date AND type = 'margin_reward' AND user_id  IN (SELECT id FROM whitelist)) AS rebate,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE truncate(create_time,86400)= @date AND type IN @activityTypes AND user_id  IN (SELECT id FROM whitelist)) AS activity,
		(SELECT COALESCE(SUM(fee_coin), 0) FROM ba_financial_withdraw WHERE truncate(update_time,86400) = @date AND status IN (1, 3) AND user_id  IN (SELECT id FROM whitelist)) AS fee,
		-(SELECT COALESCE(SUM(income), 0) FROM ba_trade_contract_order WHERE truncate(buy_time,86400) = @date AND payment_status = 1 AND user_id  IN (SELECT id FROM whitelist)) AS payment,
		0 AS bonus,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE truncate(create_time,86400) = @date AND type = 'miners_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS miners_produce,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE truncate(create_time,86400) = @date AND type = 'management_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS management_income,
		UNIX_TIMESTAMP() AS create_time
	ON DUPLICATE KEY UPDATE
		reg_nums = VALUES(reg_nums),
		income = VALUES(income),
		consumption = VALUES(consumption),
		recharge_coin = VALUES(recharge_coin),
		recharge_money = VALUES(recharge_money),
		management_buy = VALUES(management_buy),
		miners_buy = VALUES(miners_buy),
		withdraw = VALUES(withdraw),
		rebate = VALUES(rebate),
		activity = VALUES(activity),
		fee = VALUES(fee),
		payment = VALUES(payment),
		bonus = VALUES(bonus),
		miners_produce = VALUES(miners_produce),
		management_income = VALUES(management_income),
		create_time = VALUES(create_time);`
	date := time.Now().Truncate(time.Hour * 24).Unix()
	mainCoinId := rserver.GetCfgValueInt64("main_coin")
	activityTypes := []string{
		"check_in_reward", "first_contract_amount_reached", "month_invite_reached_give",
		"week_invite_reached_give", "today_invite_reached_give", "month_contract_amount_reached",
		"today_contract_amount_reached", "today_contract_num_reached", "team_num_reached_give",
		"invite_num_reached_give", "today_recharge_reached_give", "first_recharge_reached_give",
		"auth_give", "u_recharge", "invite_first_recharge", "invite_register_reward", "register_reward",
	}
	fmt.Println(" statistics report generated begin")
	err := utils.Orm.Exec(sql_stat, map[string]interface{}{
		"date":          date,
		"mainCoinId":    mainCoinId,
		"activityTypes": activityTypes,
	}).Error
	fmt.Println(" statistics report generated successfully", err)
}

/*php code
public function reportStatisticsTotal()
    {
        $mainCoinId = get_sys_config('main_coin');
        $activityTypes = ['check_in_reward', 'first_contract_amount_reached', 'month_invite_reached_give',
            'week_invite_reached_give', 'today_invite_reached_give', 'month_contract_amount_reached',
            'today_contract_amount_reached', 'today_contract_num_reached', 'team_num_reached_give',
            'invite_num_reached_give', 'today_recharge_reached_give', 'first_recharge_reached_give',
            'auth_give', 'u_recharge', 'invite_first_recharge', 'invite_register_reward', 'register_reward'];
        $whitelistUserIds = User::where('is_whitelist', 1)->column('id');
        $noWhitelistUserQuery = function (Query $query) use ($whitelistUserIds) {
            $query->where('user_id', 'not in', $whitelistUserIds);
        };
        $regNums = User::count();
        $minersIncome = CoinChange::where($noWhitelistUserQuery)->where('type', 'miners_income')->where('coin_id', $mainCoinId)->sum('amount');
        $managementIncome = CoinChange::where($noWhitelistUserQuery)->where('type', 'management_income')->where('coin_id', $mainCoinId)->sum('amount');
        $income = bcadd($minersIncome, $managementIncome, 2);
        $minersConsumption = Order::where('settlement_coin_id', $mainCoinId)->where($noWhitelistUserQuery)->sum('total_price');
        $managementConsumption = ManagementOrder::where('settlement_coin_id', $mainCoinId)->where($noWhitelistUserQuery)->sum('total_price');
        $consumption = bcadd($minersConsumption, $managementConsumption, 2);
        $rechargeMoney = FinancialRecharge::where('status', 1)->where($noWhitelistUserQuery)->sum('amount');
        $rechargeCoin = Recharge::where($noWhitelistUserQuery)->sum('amount');
        $withdraw = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '1,3')->where('type', 0)->sum('money');
        $withdrawCoin = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '1,3')->where('type', 1)->sum('coin_num');
        $rebate = CommissionChange::where($noWhitelistUserQuery)->where('type', 'margin_reward')->sum('amount');
        $activity = CoinChange::where($noWhitelistUserQuery)->where('type', 'in', $activityTypes)->sum('amount');
        $withdrawFee = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '1,3')->sum('fee_coin');
        $fee = $withdrawFee;
        $payment = ContractOrder::where($noWhitelistUserQuery)->where('payment_status', 1)->sum('income');
        $payment = abs($payment);
        $bonus = 0;
        $minersProduce = CoinChange::where($noWhitelistUserQuery)->where('type', 'miners_income')->where('coin_id', $mainCoinId)->sum('amount');
        $data = [
            'reg_nums' => $regNums,
            'income' => $income,
            'consumption' => $consumption,
            'recharge_money' => $rechargeMoney,
            'recharge_coin' => $rechargeCoin,
            'withdraw' => $withdraw,
            'withdrawCoin' => $withdrawCoin,
            'rebate' => $rebate,
            'activity' => $activity,
            'fee' => $fee,
            'payment' => $payment,
            'bonus' => $bonus,
            'miners_produce' => $minersProduce,
        ];
        RedisUtil::set('reportStatisticsTotal', $data);
        $this->success('总统计报表生成成功');
    }
*/

var sql_total = `
INSERT INTO ba_report_statistics_total (reg_nums, income, consumption, recharge_money, recharge_coin, withdraw, withdraw_coin, rebate, activity, fee, payment, bonus, miners_produce,create_time)
 select * from (
	WITH whitelist AS (
		SELECT id from ba_user where is_whitelist =0
	)
	SELECT
		(SELECT COUNT(1) FROM ba_user) AS reg_nums,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE type = 'miners_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) +
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE type = 'management_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS income,
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_miners_order WHERE settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) +
		(SELECT COALESCE(SUM(total_price), 0) FROM ba_trade_management_order WHERE settlement_coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS consumption,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_financial_recharge WHERE status = 1 AND user_id  IN (SELECT id FROM whitelist)) AS recharge_money,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_coin_recharge WHERE user_id  IN (SELECT id FROM whitelist)) AS recharge_coin,
		(SELECT COALESCE(SUM(money), 0) FROM ba_financial_withdraw WHERE status IN (1, 3) AND type = 0 AND user_id  IN (SELECT id FROM whitelist)) AS withdraw,
		(SELECT COALESCE(SUM(coin_num), 0) FROM ba_financial_withdraw WHERE status IN (1, 3) AND type = 1 AND user_id  IN (SELECT id FROM whitelist)) AS withdraw_coin,
		(SELECT COALESCE(SUM(amount), 0) FROM musk_test.ba_user_commission_change WHERE type = 'margin_reward' AND user_id  IN (SELECT id FROM whitelist)) AS rebate,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE type IN @activityTypes AND user_id  IN (SELECT id FROM whitelist)) AS activity,
		(SELECT COALESCE(SUM(fee_coin), 0) FROM ba_financial_withdraw WHERE status IN (1, 3) AND user_id  IN (SELECT id FROM whitelist)) AS fee,
		ABS((SELECT COALESCE(SUM(income), 0) FROM ba_trade_contract_order WHERE payment_status = 1 AND user_id  IN (SELECT id FROM whitelist))) AS payment,
		0 AS bonus,
		(SELECT COALESCE(SUM(amount), 0) FROM ba_user_coin_change WHERE type = 'miners_income' AND coin_id = @mainCoinId AND user_id  IN (SELECT id FROM whitelist)) AS miners_produce,
       sysdate() create_time) aa
	ON DUPLICATE KEY UPDATE
		reg_nums = aa.reg_nums,
		income = aa.income,
		consumption = aa.consumption,
		recharge_money = aa.recharge_money,
		recharge_coin = aa.recharge_coin,
		withdraw = aa.withdraw,
		withdraw_coin = aa.withdraw_coin,
		rebate = aa.rebate,
		activity = aa.activity,
		fee = aa.fee,
		payment = aa.payment,
		bonus = aa.bonus,
		miners_produce = aa.miners_produce`

func ReportStatisticsTotal() {
	workingMap.Store("ReportStatisticsTotal", true)
	defer workingMap.Delete("ReportStatisticsTotal")
	log.Println("ReportStatisticsTotal begin")
	defer log.Println("ReportStatisticsTotal end")
	date := time.Now().Format("2006-01-02")
	mainCoinId := rserver.GetCfgValueInt64("main_coin")
	activityTypes := []string{
		"check_in_reward", "first_contract_amount_reached", "month_invite_reached_give",
		"week_invite_reached_give", "today_invite_reached_give", "month_contract_amount_reached",
		"today_contract_amount_reached", "today_contract_num_reached", "team_num_reached_give",
		"invite_num_reached_give", "today_recharge_reached_give", "first_recharge_reached_give",
		"auth_give", "u_recharge", "invite_first_recharge", "invite_register_reward", "register_reward",
	}
	//var whitelistUserIds []int64
	//utils.Orm.Model(&model.User{}).Where("is_whitelist = ?", 1).Pluck("id", &whitelistUserIds)

	err := utils.Orm.Exec(sql_total, map[string]interface{}{
		"mainCoinId":    mainCoinId,
		"activityTypes": activityTypes,
	}).Error

	fmt.Println(date+" total statistics report generated successfully", err)
}

//todo reportTeamStatistics reportTeamStatisticsTotal
/* php code
public function reportTeamStatistics()
  {
      $users = User::where('id', '>', 1)->select();
      $mainCoinId = get_sys_config('main_coin');
      $whitelistUserIds = User::where('is_whitelist', 1)->column('id');
      foreach ($users as $user) {
          $userId = $user->id;
          $childIds = Db::query('select queryChildrenUsers(:refereeid) as childIds', ['refereeid' => $userId])[0]['childIds'];
          $childIds .= ',' . $userId;
          $childIds = explode(',', $childIds);
          $childIds = array_diff($childIds, $whitelistUserIds);
          $teamQuery = function (Query $query) use ($childIds) {
              $query->where('user_id', 'in', $childIds);
          };
          $assets = Assets::mainCoinAssets($userId);
          $balance = $assets->balance;
          $teamBalance = Assets::where($teamQuery)->where('coin_id', $mainCoinId)->sum('balance');
          $refereeNums = $user->referee_nums;
          $totalRefereeNums = User::where('refereeid', $userId)->where('id', 'not in', $whitelistUserIds)->count();
          $refereeNums = '总 ' . $totalRefereeNums . ' / ' . $refereeNums;
          $teamNums = $user->team_nums;
          $totalTeamNums = count($childIds) - 1; // 因为 queryChildrenUsers 这个函数返回的第一个元素是 $
          $teamNums = '总 ' . $totalTeamNums . ' / ' . $teamNums;
          $todayRechargeMoney = FinancialRecharge::where(['user_id' => $userId, 'status' => 1])->whereDay('create_time')->sum('amount');
          $totalRechargeMoney = FinancialRecharge::where(['user_id' => $userId, 'status' => 1])->sum('amount');
          $todayTeamRechargeCoin = Recharge::where($teamQuery)->whereDay('create_time')->sum('amount');
          $totalTeamRechargeCoin = Recharge::where($teamQuery)->sum('amount');
          $todayTeamRechargeMoney = FinancialRecharge::where($teamQuery)->where('status', 1)->whereDay('create_time')->sum('amount');
          $teamRechargeMoney = FinancialRecharge::where($teamQuery)->where('status', 1)->sum('amount');
          $teamWithdraw = Withdraw::where($teamQuery)->where('status', 'in', '1,3')->where('type', 0)->sum('money');
          $teamLeftWithdraw = Withdraw::where($teamQuery)->where('status', 'in', '0,4')->where('type', 0)->sum('money');
          $teamWithdrawCoin = Withdraw::where($teamQuery)->where('status', 'in', '1,3')->where('type', 1)->sum('coin_num');
          $teamLeftWithdrawCoin = Withdraw::where($teamQuery)->where('status', 'in', '0,4')->where('type', 1)->sum('coin_num');
          $data = [
              'user_id' => $userId,
              'balance' => $balance,
              'team_balance' => $teamBalance,
              'referee_nums' => $refereeNums,
              'team_nums' => $teamNums,
              'today_recharge_money' => $todayRechargeMoney,
              'total_recharge_money' => $totalRechargeMoney,
              'today_team_recharge_coin' => $todayTeamRechargeCoin,
              'total_team_recharge_coin' => $totalTeamRechargeCoin,
              'today_team_recharge_money' => $todayTeamRechargeMoney,
              'team_recharge_money' => $teamRechargeMoney,
              'team_withdraw' => $teamWithdraw,
              'team_left_withdraw' => $teamLeftWithdraw,
              'team_withdraw_coin' => $teamWithdrawCoin,
              'team_left_withdraw_coin' => $teamLeftWithdrawCoin,
          ];
          $teamStatistics = TeamStatistics::where('user_id', $userId)->find();
          if ($teamStatistics) {
              $teamStatistics->save($data);
          } else {
              TeamStatistics::create($data);
          }
      }
      $this->success('团队报表的数据生成完毕');
  }
*/
func ReportTeamStatistics() {
	workingMap.Store("ReportTeamStatistics", true)
	defer workingMap.Delete("ReportTeamStatistics")
	log.Println("ReportTeamStatistics begin")
	defer log.Println("ReportTeamStatistics end")

	sql := `
INSERT INTO ba_report_team_statistics (
    user_id, balance, team_balance, referee_nums, team_nums,
    today_recharge_money, total_recharge_money, total_team_recharge_coin,today_team_recharge_coin
    ,team_recharge_money,today_team_recharge_money,
    team_withdraw, team_left_withdraw, team_withdraw_coin, team_left_withdraw_coin,update_time
)
WITH team_stats AS (
    SELECT
        pid,
        COUNT(1) AS total_team_nums,
        SUM(CASE WHEN team_level = 1 THEN 1 ELSE 0 END) AS total_referee_nums,
        SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END) AS team_nums,
        SUM(CASE WHEN team_level = 1 AND is_active = 1 THEN 1 ELSE 0 END) AS referee_nums
    FROM ba_team_user
    WHERE is_whitelist = 0
    GROUP BY pid
),
leader_stats as(
    select pid,COALESCE( sum(fr.amount),0) AS total_recharge_money,
        COALESCE(SUM(CASE WHEN DATE(FROM_UNIXTIME(fr.create_time)) = CURDATE()
            THEN fr.amount ELSE 0 END), 0) AS today_recharge_money
    from team_stats ts inner join ba_financial_recharge fr on ts.pid=fr.user_id and fr.status = 1
     GROUP BY pid
),
team_query AS (
    WITH fr_agg AS (
    SELECT
        user_id,
        COALESCE(SUM(amount), 0) AS total_team_recharge_money,
        COALESCE(SUM(CASE WHEN DATE(FROM_UNIXTIME(create_time)) = CURDATE() THEN amount ELSE 0 END), 0) AS today_team_recharge_money
    FROM ba_financial_recharge
    WHERE status = 1
    GROUP BY user_id
    ),
    r_agg AS (
        SELECT
            user_id,
            COALESCE(SUM(amount), 0) AS total_team_recharge_coin,
            COALESCE(SUM(CASE WHEN DATE(FROM_UNIXTIME(create_time)) = CURDATE() THEN amount ELSE 0 END), 0) AS today_team_recharge_coin
        FROM ba_coin_recharge
        GROUP BY user_id
    ),
    w_agg AS (
        SELECT
            user_id,
            COALESCE(SUM(CASE WHEN status IN (1, 3) AND type = 0 THEN money ELSE 0 END), 0) AS team_withdraw,
            COALESCE(SUM(CASE WHEN status IN (0, 4) AND type = 0 THEN money ELSE 0 END), 0) AS team_left_withdraw,
            COALESCE(SUM(CASE WHEN status IN (1, 3) AND type = 1 THEN coin_num ELSE 0 END), 0) AS team_withdraw_coin,
            COALESCE(SUM(CASE WHEN status IN (0, 4) AND type = 1 THEN coin_num ELSE 0 END), 0) AS team_left_withdraw_coin
        FROM ba_financial_withdraw
        GROUP BY user_id
    ),
    whole_team as(
        select id,pid from ba_team_user tu WHERE tu.is_whitelist = 0
        union all
        select id,id pid  from ba_user u inner join
        (select distinct pid  from  ba_team_user) leader on u.id=leader.pid and u.is_whitelist=0
    )
        SELECT
            tu.pid,
            COALESCE(SUM(a.balance), 0) AS team_balance,
            COALESCE(r_agg.total_team_recharge_coin, 0) AS total_team_recharge_coin,
            COALESCE(r_agg.today_team_recharge_coin, 0) AS today_team_recharge_coin,
            COALESCE(fr_agg.total_team_recharge_money, 0) AS total_team_recharge_money,
            COALESCE(fr_agg.today_team_recharge_money, 0) AS today_team_recharge_money,
            COALESCE(w_agg.team_withdraw, 0) AS team_withdraw,
            COALESCE(w_agg.team_left_withdraw, 0) AS team_left_withdraw,
            COALESCE(w_agg.team_withdraw_coin, 0) AS team_withdraw_coin,
            COALESCE(w_agg.team_left_withdraw_coin, 0) AS team_left_withdraw_coin
        FROM whole_team tu
        LEFT JOIN fr_agg ON tu.id = fr_agg.user_id
        LEFT JOIN r_agg ON tu.id = r_agg.user_id
        LEFT JOIN w_agg ON tu.id = w_agg.user_id
        LEFT JOIN ba_user_assets a ON tu.id = a.user_id and a.coin_id=1
        GROUP BY tu.pid
)
SELECT
    ts.pid AS user_id,
    COALESCE(a.balance,0) balance,
    tq.team_balance,
    CONCAT('总 ', ts.total_referee_nums, ' / ', ts.referee_nums) AS referee_nums,
    CONCAT('总 ', ts.total_team_nums, ' / ', ts.team_nums) AS team_nums,
    ifnull(us.total_recharge_money,0) total_recharge_money,
     ifnull(us.today_recharge_money,0) today_recharge_money,
     ifnull(tq.total_team_recharge_coin,0) total_team_recharge_coin,
     ifnull(tq.today_team_recharge_coin,0) today_team_recharge_coin,
     ifnull(tq.total_team_recharge_money ,0) team_recharge_money,
     ifnull(tq.today_team_recharge_money,0) today_team_recharge_money,
     ifnull(tq.team_withdraw,0) team_withdraw,
     ifnull(tq.team_left_withdraw,0) team_left_withdraw,
     ifnull(tq.team_withdraw_coin,0) team_withdraw_coin,
     ifnull(tq.team_left_withdraw_coin,0) team_left_withdraw_coin,
    unix_timestamp() update_time
FROM
    team_stats ts
    left join leader_stats us on ts.pid=us.pid
    LEFT JOIN ba_user_assets a ON ts.pid = a.user_id AND  a.coin_id=1
    LEFT JOIN team_query tq ON ts.pid = tq.pid
ON DUPLICATE KEY UPDATE
    user_id=values(user_id),
    balance = VALUES(balance),
    team_balance = VALUES(team_balance),
    referee_nums = VALUES(referee_nums),
    team_nums = VALUES(team_nums),
    today_recharge_money = VALUES(today_recharge_money),
    total_recharge_money = VALUES(total_recharge_money),
    total_team_recharge_coin = VALUES(total_team_recharge_coin),
    today_team_recharge_coin = VALUES(today_team_recharge_coin),
    team_recharge_money = VALUES(team_recharge_money),
    today_team_recharge_money = VALUES(today_team_recharge_money),
    team_withdraw = VALUES(team_withdraw),
    team_left_withdraw = VALUES(team_left_withdraw),
    team_withdraw_coin = VALUES(team_withdraw_coin),
    team_left_withdraw_coin = VALUES(team_left_withdraw_coin),
    update_time=values(update_time);
`

	err := utils.Orm.Exec(sql).Error
	if err != nil {
		log.Println("Error generating team statistics:", err)
		return
	}

	log.Println("Team statistics generated successfully")
}

/*
	php code

public function reportTeamStatisticsTotal()

	{
	    $mainCoinId = get_sys_config('main_coin');
	    $noWhitelistUserQuery = User::where('is_whitelist', 0)->column('id');
	    $todayRegNums = User::whereDay('create_time')->where('id', 'not in', $whitelistUserIds)->count();
	    $totalRegNums = User::count();
	    $todayRechargeCoin = Recharge::where($noWhitelistUserQuery)->whereDay('create_time')->sum('amount');
	    $todayRechargeMoney = FinancialRecharge::where($noWhitelistUserQuery)->where('status', 1)->whereDay('create_time')->sum('amount');
	    $totalWithdrawMoney = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '1,3')->where('type', 0)->sum('money');
	    $leftWithdrawMoney = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '0,4')->where('type', 0)->sum('money');
	    $totalWithdrawCoin = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '1,3')->where('type', 1)->sum('coin_num');
	    $leftWithdrawCoin = Withdraw::where($noWhitelistUserQuery)->where('status', 'in', '0,4')->where('type', 1)->sum('coin_num');
	    $totalBalance = Assets::where($noWhitelistUserQuery)->where('coin_id', $mainCoinId)->sum('balance');
	    $totalRechargeCoin = Recharge::where($noWhitelistUserQuery)->sum('amount');
	    $totalRechargeMoney = FinancialRecharge::where($noWhitelistUserQuery)->where('status', 1)->sum('amount');;
	    $data = [
	        'todayRegNums' => $todayRegNums,
	        'totalRegNums' => $totalRegNums,
	        'todayRechargeCoin' => $todayRechargeCoin,
	        'todayRechargeMoney' => $todayRechargeMoney,
	        'totalWithdrawMoney' => $totalWithdrawMoney,
	        'leftWithdrawMoney' => $leftWithdrawMoney,
	        'totalWithdrawCoin' => $totalWithdrawCoin,
	        'leftWithdrawCoin' => $leftWithdrawCoin,
	        'totalBalance' => $totalBalance,
	        'totalRechargeCoin' => $totalRechargeCoin,
	        'totalRechargeMoney' => $totalRechargeMoney,
	    ];
	    RedisUtil::set('reportTeamStatisticsTotal', $data);
	    $this->success('团队总的报表数据生成成功');
	}
*/
func ReportTeamStatisticsTotal() {
	workingMap.Store("ReportTeamStatisticsTotal", true)
	defer workingMap.Delete("ReportTeamStatisticsTotal")
	log.Println("ReportTeamStatisticsTotal begin")
	defer log.Println("ReportTeamStatisticsTotal end")

	sql := `
    INSERT INTO ba_report_team_statistics_total (
        today_reg_nums, total_reg_nums, today_recharge_coin, today_recharge_money,
        total_withdraw_money, left_withdraw_money, total_withdraw_coin, left_withdraw_coin,
        total_balance, total_recharge_coin, total_recharge_money, create_time
    )
    SELECT
        (SELECT COUNT(*) FROM ba_user WHERE DATE(from_unixtime(create_time)) = CURDATE() AND is_whitelist = 0) AS today_reg_nums,
        (SELECT COUNT(*) FROM ba_user) AS total_reg_nums,
        (SELECT COALESCE(SUM(amount), 0) FROM ba_coin_recharge WHERE DATE(from_unixtime(create_time)) = CURDATE() AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS today_recharge_coin,
        (SELECT COALESCE(SUM(amount), 0) FROM ba_financial_recharge WHERE status = 1 AND DATE(from_unixtime(create_time)) = CURDATE() AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS today_recharge_money,
        (SELECT COALESCE(SUM(money), 0) FROM ba_financial_withdraw WHERE status IN (1, 3) AND type = 0 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS total_withdraw_money,
        (SELECT COALESCE(SUM(money), 0) FROM ba_financial_withdraw WHERE status IN (0, 4) AND type = 0 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS left_withdraw_money,
        (SELECT COALESCE(SUM(coin_num), 0) FROM ba_financial_withdraw WHERE status IN (1, 3) AND type = 1 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS total_withdraw_coin,
        (SELECT COALESCE(SUM(coin_num), 0) FROM ba_financial_withdraw WHERE status IN (0, 4) AND type = 1 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS left_withdraw_coin,
        (SELECT COALESCE(SUM(balance), 0) FROM ba_user_assets WHERE coin_id = 1 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS total_balance,
        (SELECT COALESCE(SUM(amount), 0) FROM ba_coin_recharge WHERE user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS total_recharge_coin,
        (SELECT COALESCE(SUM(amount), 0) FROM ba_financial_recharge WHERE status = 1 AND user_id IN (SELECT id FROM ba_user WHERE is_whitelist = 0)) AS total_recharge_money,
        UNIX_TIMESTAMP() create_time
    ON DUPLICATE KEY UPDATE
        today_reg_nums = VALUES(today_reg_nums),
        total_reg_nums = VALUES(total_reg_nums),
        today_recharge_coin = VALUES(today_recharge_coin),
        today_recharge_money = VALUES(today_recharge_money),
        total_withdraw_money = VALUES(total_withdraw_money),
        left_withdraw_money = VALUES(left_withdraw_money),
        total_withdraw_coin = VALUES(total_withdraw_coin),
        left_withdraw_coin = VALUES(left_withdraw_coin),
        total_balance = VALUES(total_balance),
        total_recharge_coin = VALUES(total_recharge_coin),
        total_recharge_money = VALUES(total_recharge_money),
        create_time = VALUES(create_time);
    `

	err := utils.Orm.Exec(sql).Error
	if err != nil {
		log.Println("Error generating team statistics total:", err)
		return
	}

	log.Println("Team statistics total generated successfully")
}
