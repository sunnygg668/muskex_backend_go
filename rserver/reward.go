package rserver

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"sort"
	"strconv"
	"time"
)

type Reward struct {
}

/*
php code:
public function userActivation(Job $job, $data)

	{
	    $userId = $data['user_id'];
	    Db::startTrans();
	    try {
	        $user = User::find($userId);
	        $oldIsActivation = $user->is_activation;
	        if ($user->refereeid) {
	            $refereeUser = User::find($user->refereeid);
	            if ($refereeUser) {
	                $inviteOneGiveCoinType = get_sys_config('invite_one_give_coin_type');
	                $inviteOneGiveCoinNum = get_sys_config('invite_one_give_coin_num');
	                $coinChange = CoinChange::getRecordFrom($refereeUser->id, $userId, 'invite_register_reward');
	                if (!$coinChange && $inviteOneGiveCoinType && $inviteOneGiveCoinNum > 0) {
	                    Assets::updateCoinAssetsBalance($refereeUser->id, $inviteOneGiveCoinType, $inviteOneGiveCoinNum, 'invite_register_reward', $userId);
	                }
	            }

	        }
	        Db::commit();
	        $job->delete();
	        if ($oldIsActivation == 0) {
	            $user->save(['is_activation' => 1, 'activation_time' => time()]);
	            Queue::push('\app\custom\job\UserQueue@updateTeamLevel', ['user_id' => $userId, 'num' => 1], 'user');
	        }
	    } catch (\Exception $e) {
	        Db::rollback();
	        Log::error($e->getMessage());
	    }
	}
*/

func (r Reward) userActive(user *model.User) {
	if user.IsActivation == 0 {
		user.IsActivation = 1
		user.ActiveTimes += 1
		utils.Orm.Updates(&model.User{Id: user.Id, IsActivation: 1, ActivationTime: time.Now().Unix(),
			ActiveTimes: user.ActiveTimes})
		r.UpdateTeamLevelFroActive(user.Id, true)
		if user.Refereeid != 0 && user.ActiveTimes == 1 {
			refereeUser := new(model.User)
			utils.Orm.First(refereeUser, user.Refereeid)
			if refereeUser.Id > 0 {
				inviteOneGiveCoinType, _ := strconv.Atoi(GetCfgValue("invite_one_give_coin_type"))
				inviteOneGiveCoinNum, _ := strconv.ParseFloat(GetCfgValue("invite_one_give_coin_num"), 64)
				coinChanged := int64(0)
				utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND from_user_id=? AND type =?", user.Refereeid, user.Id, "invite_register_reward").Count(&coinChanged)
				if coinChanged > 0 && inviteOneGiveCoinType != 0 && inviteOneGiveCoinNum > 0 {
					UpdateCoinAssetsBalance(utils.Orm, user.Refereeid, int64(inviteOneGiveCoinType), inviteOneGiveCoinNum, "invite_register_reward", 0, user.Id, "")
				}
			}
		}

	}
}
func (Reward) UpdateTeamLevelFroActive(userId int64, isActive bool) {
	utils.Orm.Where("id = ?", userId).Select("IsActive").Updates(&model.TeamUser{IsActive: isActive})
	if isActive {
		rew := reward1{}
		rew.inviteGive(userId)
	}
}

/*
php code:

	public function userFirstRecharge(Job $job, $data)
	   {
	       $userId = $data['user_id'];
	       $user = User::find($userId);
	       if ($user->refereeid) {
	           $refereeUser = User::find($user->refereeid);
	           if ($refereeUser) {
	               $inviteRechargeGiveCoinType = get_sys_config('invite_recharge_give_coin_type');
	               $inviteRechargeGiveCoinNum = get_sys_config('invite_recharge_give_coin_num');
	               $coinChange = CoinChange::getRecordFrom($refereeUser->id, $userId, 'invite_first_recharge');
	               if (!$coinChange && $inviteRechargeGiveCoinType && $inviteRechargeGiveCoinNum > 0) {
	                   Assets::updateCoinAssetsBalance($refereeUser->id, $inviteRechargeGiveCoinType, $inviteRechargeGiveCoinNum, 'invite_first_recharge', $userId);
	               }
	           }
	       }
	       $job->delete();
	   }
*/
func (Reward) userFirstRecharge(user *model.User, orderid int64) {
	if user.Refereeid != 0 {
		refereeUser := new(model.User)
		utils.Orm.First(refereeUser, user.Refereeid)
		if refereeUser.Id != 0 {
			inviteRechargeGiveCoinType := GetCfgValueInt64("invite_recharge_give_coin_type")
			inviteRechargeGiveCoinNum, _ := strconv.ParseFloat(GetCfgValue("invite_recharge_give_coin_num"), 64)
			coinChanged := int64(0)
			utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND from_user_id=? AND type =?", user.Refereeid, user.Id, "invite_first_recharge").Count(&coinChanged)
			if coinChanged == 0 && inviteRechargeGiveCoinType != 0 && inviteRechargeGiveCoinNum > 0 {
				UpdateCoinAssetsBalance(utils.Orm, user.Refereeid, int64(inviteRechargeGiveCoinType), inviteRechargeGiveCoinNum, "invite_first_recharge", orderid, user.Id, "")
			}
		}
	}
}

/*
php code:
public function contractBuy(Job $job, $data)

	{
	    $userId = $data['user_id'];
	    $margin = $data['margin'];
	    $user = User::find($userId);
	    $levelArray = [];
	    $levelList = Level::where('is_open', 1)->select();
	    if (empty($levelList)) {
	        $job->delete();
	        return;
	    }
	    foreach ($levelList as $level) {
	        $levelArray[$level->level] = $level;
	    }
	    Db::startTrans();
	    try {
	        $maxRebateLayers = Level::where('is_open', 1)->max('rebate_layers');
	        for ($i = 1; $i <= $maxRebateLayers; $i++) {
	            if (!$user->refereeid) {
	                break;
	            }
	            $refereeUser = User::find($user->refereeid);
	            if (!$refereeUser) {
	                break;
	            }
	            $level = $levelArray[$refereeUser->level];
	            if ($level->rebate_layers < $i) {
	                continue;
	            }
	            $ratioKey = 'layer_' . $i . '_ratio';
	            if (empty($level[$ratioKey])) {
	                continue;
	            }
	            $ratio = $level[$ratioKey];
	            $commission = bcmul($margin, $ratio / 100, 2);
	            User::updateCommission($refereeUser->id, $commission, 'margin_reward', $userId);
	            $user = $refereeUser;
	        }
	        Db::commit();
	        $job->delete();
	    } catch (\Exception $e) {
	        Db::rollback();
	        Log::error($e->getMessage());
	    }
	}
*/
func (Reward) contractBuyForParent(user *model.User, margin float64, orderid int64) {
	levelMap := map[uint32]*model.UserLevel{}
	levelList := []*model.UserLevel{}
	utils.Orm.Order("rebate_layers desc").Where("is_open", 1).Find(&levelList)
	if len(levelList) == 0 {
		return
	}
	for _, level := range levelList {
		levelMap[level.Level] = level
	}

	maxRebateLayers := int(levelList[0].RebateLayers)
	for i := 1; i <= maxRebateLayers; i++ {
		if user.Refereeid == 0 {
			break
		}
		refereeUser := new(model.User)
		utils.Orm.First(refereeUser, user.Refereeid)
		if refereeUser.Id == 0 {
			break
		}
		level := levelMap[uint32(refereeUser.Level)]
		if level.RebateLayers < int64(i) {
			continue
		}
		ratioKey := "layer_" + strconv.Itoa(i) + "_ratio"
		ratio, _ := strconv.ParseFloat(GetCfgValue(ratioKey), 64)
		if ratio == 0 {
			continue
		}
		commission := margin * (ratio / 100)
		updateCommission(refereeUser, commission, "margin_reward", strconv.Itoa(int(user.Id)), orderid)
		user = refereeUser
	}

}

/*
php code:

	public function managementBuy(Job $job, $data)
	{
	    $userId = $data['user_id'];
	    $rebate_income = $data['rebate_income'];

	    Db::startTrans();
	    try {
	        $commission = $rebate_income;
	        User::updateCommission($userId, $commission, 'rebate_income', $userId);
	        Db::commit();
	        $job->delete();
	    } catch (\Exception $e) {
	        Db::rollback();
	        Log::error($e->getMessage());
	    }
	}
*/
func (Reward) managementBuy(user *model.User, rebate_income float64) {
	commission := rebate_income
	updateCommission(user, commission, "rebate_income", "", 0)
}

/*
php code:

	public function minersLease(Job $job, $data)
	   {
	       $userId = $data['user_id'];
	       $totalPrice = $data['totalPrice'];
	       $user = User::find($userId);
	       $giveMinersRewardLevel = get_sys_config('give_miners_reward_level');
	       Db::startTrans();
	       try {
	           for ($i = 1; $i <= 2; $i++) {
	               if (!$user->refereeid) {
	                   break;
	               }
	               $refereeUser = User::find($user->refereeid);
	               if (!$refereeUser) {
	                   break;
	               }
	               $ratioKey = 'layer_' . $i . '_miners_reward_ratio';
	               $ratio = get_sys_config($ratioKey);
	               if ($ratio) {
	                   $amount = bcmul($totalPrice, $ratio / 100, 2);
	                   Assets::updateMainCoinAssetsBalance($refereeUser->id, $amount, 'miners_reward', $userId);
	               }
	               $user = $refereeUser;
	           }
	           Db::commit();
	           $job->delete();
	       } catch (\Exception $e) {
	           Db::rollback();
	           Log::error($e->getMessage());
	       }
	   }
*/
func (Reward) minersLease(user *model.User, totalPrice float64, orderId int64) {
	tx := utils.Orm.Begin()
	defer tx.Rollback()

	for i := 1; i <= 2; i++ {
		if user.Refereeid == 0 {
			break
		}
		refereeUser := new(model.User)
		utils.Orm.First(refereeUser, user.Refereeid)
		if refereeUser.Id > 0 {
			break
		}
		ratioKey := "layer_" + strconv.Itoa(i) + "_miners_reward_ratio"
		ratio, _ := strconv.ParseFloat(GetCfgValue(ratioKey), 64)
		if ratio != 0 {
			amount := totalPrice * (ratio / 100)
			UpdateCoinAssetsBalance(utils.Orm, refereeUser.Id, 1, amount, "miners_reward", orderId, user.Id, "")
		}
		user = refereeUser
	}
	tx.Commit()
}

/*
php code:

	public function giveLotteryCount(Job $job, $data)
	{
	    $userId = $data['user_id'];
	    $coinAmount = $data['coinAmount'];
	    $mainCoinPrice = Assets::mainCoinPrice();
	    $amount = bcmul($coinAmount, $mainCoinPrice, 2);
	    $rechargeGiveCount = get_sys_config('recharge_give_count');
	    if ($rechargeGiveCount) {
	        array_multisort(array_column($rechargeGiveCount, 'key'), SORT_DESC, $rechargeGiveCount);
	        foreach ($rechargeGiveCount as $giveCount) {
	            if ($amount >= $giveCount['key'] && $giveCount['value'] > 0) {
	                User::where('id', $userId)->inc('lottery_count', $giveCount['value']);
	                break;
	            }
	        }
	    }
	    $job->delete();
	}
*/
func (Reward) giveLotteryCount(user *model.User, coinAmount float64) {
	mainCoinPrice := getUPrice()
	amount := coinAmount * mainCoinPrice
	grades := collectGrades{}
	grades.Process("recharge_give_count", amount, func(key, value float64) {
		utils.Orm.Model(&model.User{}).Where("id = ?", user.Id).UpdateColumn("lottery_count", gorm.Expr("lottery_count + ?", value))
	})
}

type reward1 struct {
}

func (reward1) authGive(userId int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	authGiveCoinNum, _ := strconv.ParseFloat(GetCfgValue("auth_give_coin_num"), 64)
	if authGiveCoinNum <= 0 {
		return
	}
	coinChange := new(model.UserCoinChange)
	utils.Orm.Where("user_id = ? AND type = ?", userId, "auth_give").First(coinChange)
	if coinChange.Id == 0 {
		UpdateCoinAssetsBalance(utils.Orm, userId, int64(taskGiveCoinType), authGiveCoinNum, "auth_give", 0, 0, "")
	}

}

func (reward1) firstRechargeReachedGive(userId int64, amount float64) {
	rechargeCount := int64(0)
	utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND type IN ?", userId, []string{"financial_recharge"}).Count(&rechargeCount)
	rechargeGiveCount := int64(0)
	utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND type = ?", userId, "first_recharge_reached_give").Count(&rechargeGiveCount)
	if rechargeCount > 1 || rechargeGiveCount > 0 {
		return
	}
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	grades := collectGrades{}
	grades.Process("first_recharge_reached_give", amount, func(key, value float64) {
		UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "first_recharge_reached_give", 0, 0, fmt.Sprint(key))
	})
}

func (reward1) todayRechargeReachedGive(userId int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	todayAmount := float64(0)
	utils.Orm.Model(&model.FinancialRecharge{}).Where("user_id = ? AND status IN ?", userId, []string{"1", "3"}).Where("DATE(from_unixtime(create_time)) = CURDATE()").Select("SUM(amount)").Scan(&todayAmount)
	grades := collectGrades{}
	grades.Process("today_recharge_reached_give", todayAmount, func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND DATE(from_unixtime(create_time)) = CURDATE()", userId, "today_recharge_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, int64(taskGiveCoinType), value, "today_recharge_reached_give", 0, 0, "")
		}
	})
}

func (r reward1) inviteGive(userId int64) {
	r.inviteNumReachedGive(userId)
	r.todayInviteReachedGive(userId)
	r.weekInviteReachedGive(userId)
	r.monthInviteReachedGive(userId)
	r.inviterTeamNumReachedGive(userId)
}
func (reward1) inviteNumReachedGive(userId int64) {
	user := new(model.User)
	if err := utils.Orm.First(user, userId).Error; err != nil {
		return
	}
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	refereeNums := user.RefereeNums
	grades := collectGrades{}
	grades.Process("invite_num_reached_give", float64(refereeNums), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ?", userId, "invite_num_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, int64(taskGiveCoinType), value, "invite_num_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) todayInviteReachedGive(userId int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	giveArray := []*collectGrade{}
	json.Unmarshal([]byte(GetCfgValue("today_invite_reached_give")), &giveArray)
	sort.Slice(giveArray, func(i, j int) bool {
		return giveArray[i].Key > giveArray[j].Key
	})
	todayInviteNums := int64(0)
	utils.Orm.Model(&model.User{}).Where("refereeid = ? AND is_activation = 1", userId).Where("DATE(from_unixtime(create_time)) = CURDATE()").Count(&todayInviteNums)
	grades := collectGrades{}
	grades.Process("today_invite_reached_give", float64(todayInviteNums), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND DATE(from_unixtime(create_time)) = CURDATE()", userId, "today_invite_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, int64(taskGiveCoinType), value, "today_invite_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) weekInviteReachedGive(userId int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	weekInviteNums := int64(0)
	utils.Orm.Model(&model.User{}).Where("refereeid = ? AND is_activation = 1", userId).Where("YEARWEEK(from_unixtime(create_time), 1) = YEARWEEK(CURDATE(), 1)").Count(&weekInviteNums)

	grades := collectGrades{}
	grades.Process("week_invite_reached_give", float64(weekInviteNums), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND YEARWEEK(from_unixtime(create_time), 1) = YEARWEEK(CURDATE(), 1)", userId, "week_invite_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, int64(taskGiveCoinType), value, "week_invite_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) monthInviteReachedGive(userId int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	//giveArray := []*CollectGrade{}
	//json.Unmarshal([]byte(GetCfgValue("month_invite_reached_give")), &giveArray)
	//sort.Slice(giveArray, func(i, j int) bool {
	//	return giveArray[i].Key > giveArray[j].Key
	//})
	monthInviteNums := int64(0)
	utils.Orm.Model(&model.User{}).Where("refereeid = ? AND is_activation = 1", userId).Where("MONTH(from_unixtime(create_time)) = MONTH(CURDATE())").Count(&monthInviteNums)
	grades := collectGrades{}
	grades.Process("month_invite_reached_give", float64(monthInviteNums), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND MONTH(from_unixtime(create_time)) = MONTH(CURDATE())", userId, "month_invite_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "month_invite_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}
func (reward1) inviterTeamNumReachedGive(userId int64) {
	//user := new(model.User)
	//if err := utils.Orm.First(user, userId).Error; err != nil {
	//	return
	//}
	//teamNums := user.TeamNums
	teamNums := int64(0)
	utils.Orm.Model(&model.TeamUser{}).Where("pid = ? AND is_active = 1", userId).Count(&teamNums)
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	grades := collectGrades{}
	grades.Process("team_num_reached_give", float64(teamNums), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ?", userId, "team_num_reached_give", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "team_num_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) todayContractNumReached(userId int64, orderid int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	todayContractNum := int64(0)
	utils.Orm.Model(&model.TradeContractOrder{}).Where("user_id = ?", userId).Where("DATE(from_unixtime(buy_time)) = CURDATE()").Count(&todayContractNum)
	grades := collectGrades{}
	grades.Process("today_contract_num_reached", float64(todayContractNum), func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND DATE(from_unixtime(create_time)) = CURDATE()", userId, "today_contract_num_reached", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "today_contract_num_reached", 0, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) todayContractAmountReached(userId int64, orderid int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	todayInvestedCoinNum := float64(0)
	utils.Orm.Model(&model.TradeContractOrder{}).Where("user_id = ?", userId).Where("DATE(from_unixtime(buy_time)) = CURDATE()").Select("SUM(invested_coin_num)").Scan(&todayInvestedCoinNum)
	grades := collectGrades{}
	grades.Process("today_contract_amount_reached", todayInvestedCoinNum, func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND DATE(from_unixtime(create_time)) = CURDATE()", userId, "today_contract_amount_reached", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "today_contract_amount_reached", orderid, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) monthContractAmountReached(userId, orderid int64) {
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	monthInvestedCoinNum := float64(0)
	utils.Orm.Model(&model.TradeContractOrder{}).Where("user_id = ?", userId).Where("MONTH(from_unixtime(buy_time)) = MONTH(CURDATE())").Select("SUM(invested_coin_num)").Scan(&monthInvestedCoinNum)
	grades := collectGrades{}
	grades.Process("month_contract_amount_reached", monthInvestedCoinNum, func(key, value float64) {
		coinChange := new(model.UserCoinChange)
		utils.Orm.Where("user_id = ? AND type = ? AND remark = ? AND MONTH(from_unixtime(create_time)) = MONTH(CURDATE())", userId, "month_contract_amount_reached", key).First(coinChange)
		if coinChange.Id == 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "month_contract_amount_reached", orderid, 0, fmt.Sprint(key))
		}
	})
}

func (reward1) firstContractAmountReached(userId int64, orderid int64) {
	coinChange := new(model.UserCoinChange)
	if err := utils.Orm.Where("user_id = ? AND type = ?", userId, "first_contract_amount_reached").First(coinChange).Error; err == nil {
		return
	}
	taskGiveCoinType := GetCfgValueInt64("task_give_coin_type")
	totalInvestedCoinNum := float64(0)
	utils.Orm.Model(&model.TradeContractOrder{}).Where("user_id = ?", userId).Select("SUM(invested_coin_num)").Scan(&totalInvestedCoinNum)
	grades := collectGrades{}
	grades.Process("first_contract_amount_reached", totalInvestedCoinNum, func(key, value float64) {
		UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "first_contract_amount_reached", orderid, 0, fmt.Sprint(key))
	})
}

/*
php code:
public function collect()

	{
	    $user = $this->auth->getUser();
	    $commissionPool = $user->commission_pool;
	    $teamLevel = get_sys_config('team_level');
	    $teamLevelNums = get_sys_config('team_level_nums');
	    $tl = TeamLevel::where(['user_id' => $user->id, 'user_level' => $teamLevel])->find();
	    if ($tl->team_nums < $teamLevelNums) {
	        $level = Level::where(['level' => $teamLevel])->find();
	        $this->error('团队 ' . $level->name . ' 等级的有效人数不满 ' . $teamLevelNums . ' 人');
	    }
	    $teamTotalRecharge = get_sys_config('team_total_recharge');
	    $childIds = Db::query('select queryChildrenUsers(:refereeid) as childIds', ['refereeid' => $user->id])[0]['childIds'];
	    $rechargeCoin = Recharge::where('user_id', 'in', $childIds)->sum('amount');
	    $rechargeMoney = FinancialRecharge::where('user_id', 'in', $childIds)->where('status', 1)->sum('main_coin_num');
	    $totalRecharge = bcadd($rechargeCoin, $rechargeMoney, 2);
	    if ($totalRecharge < $teamTotalRecharge) {
	        $mainCoin = Coin::mainCoin();
	        $this->error('团队总充值不满 ' . $teamTotalRecharge . ' ' . $mainCoin->name);
	    }
	    $collectNum = 0;
	    $gradeIndex = 0;
	    $teamNums = $user->team_nums;
	    $canCollect = false;
	    $collectGradeArray = get_sys_config('collect_grade');
	    array_multisort(array_column($collectGradeArray,'key'),SORT_ASC, $collectGradeArray);
	    foreach ($collectGradeArray as $key => $collectGrade) {
	        $canCollect = CommissionChange::where(['user_id' => $user->id, 'type' => 'commission_pool_collect', 'remark' => $key])->count() == 0;
	        if ($teamNums <= $collectGrade['key'] || $canCollect) {
	            $collectNum = $collectGrade['value'];
	            $gradeIndex = $key;
	            break;
	        }
	    }
	    if ($collectNum == 0 || !$canCollect) {
	        $this->error('暂不满足领取条件');
	    }
	    if ($commissionPool < $collectNum) {
	        $this->error('佣金池余额不足');
	    }
	    Db::startTrans();
	    try {
	        Assets::updateMainCoinAssetsBalance($user->id, $collectNum, 'commission_pool_collect');
	        User::updateCommission($user->id, -$collectNum, 'commission_pool_collect', null, $gradeIndex);
	        Db::commit();
	    } catch (\Exception $e) {
	        Db::rollback();
	        $this->error($e->getMessage());
	    }
	    $this->success('领取成功');
	}
*/

type teamCount struct {
	TeamNums      int64
	TeamLevelNums int
}

func (Reward) collect(user *model.User) error {
	commissionPool := user.CommissionPool
	teamLevel := GetCfgValueInt64("team_level")
	teamLevelNums := GetCfgValueInt64("team_level_nums")

	tc := new(teamCount)
	sql := `
select count(1) team_nums,
   sum(case when user_level = ? then 1 else 0 end) team_level_nums
from ba_team_user
where pid = ? and is_active=1;`
	err := utils.Orm.Raw(sql, teamLevel, user.Id).Scan(tc).Error
	if err != nil {
		return err
	}
	user.TeamNums = tc.TeamNums
	//tl := new(model.UserTeamLevel)
	//if err := utils.Orm.Where("user_id = ? AND user_level = ?", user.Id, teamLevel).First(tl).Error; err != nil {
	//	return err
	//}
	if int64(tc.TeamNums) < teamLevelNums {
		level := new(model.UserLevel)
		if err := utils.Orm.Where("level = ?", teamLevel).First(level).Error; err != nil {
			return err
		}
		return fmt.Errorf("团队 %s 等级的有效人数不满 %d", level.Name, teamLevelNums)
	}

	teamTotalRecharge := GetCfgValueF64("team_total_recharge")
	totalRecharge := float64(0)
	sql1 := `
SELECT SUM(uc.amount) + SUM(fr.main_coin_num)
FROM ba_team_user tu
LEFT JOIN ba_user_coin_change uc ON tu.id = uc.user_id
LEFT JOIN ba_financial_recharge fr ON tu.id = fr.user_id AND fr.status = 1
WHERE tu.pid = ?`
	if err := utils.Orm.Raw(sql1, user.Id).Scan(&totalRecharge).Error; err != nil {
		return err
	}

	if totalRecharge < teamTotalRecharge {
		//mainCoin := new(model.Coin)
		//if err := utils.Orm.Where("is_main = 1").First(mainCoin).Error; err != nil {
		//	return err
		//}
		return fmt.Errorf("团队总充值不满 %f %s", teamTotalRecharge, "USDT")
	}

	collectNum := float64(0)
	gradeIndex := 0
	teamNums := float64(user.TeamNums)
	canCollect := false
	grades := collectGrades{}
	grades.load("collect_grade", false)
	for _, grade := range grades {
		count := int64(0)
		err := utils.Orm.Model(&model.UserCommissionChange{}).Where("user_id = ? AND type = ? AND remark = ?", user.Id, "commission_pool_collect", grade.Key).Count(&count).Error
		if err != nil {
			return err
		}
		canCollect = count == 0
		if teamNums <= grade.Key || canCollect {
			collectNum = grade.Value
			gradeIndex = int(grade.Key)
			break
		}
	}

	if collectNum == 0 || !canCollect {
		return fmt.Errorf("暂不满足领取条件")
	}
	if commissionPool < collectNum {
		return fmt.Errorf("佣金池余额不足")
	}

	tx := utils.Orm.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := UpdateCoinAssetsBalance(tx, user.Id, 1, collectNum, "commission_pool_collect", 0, 0, ""); err != nil {
		tx.Rollback()
		return err
	}
	if err := updateCommission(user, -collectNum, "commission_pool_collect", strconv.Itoa(gradeIndex), 0); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
