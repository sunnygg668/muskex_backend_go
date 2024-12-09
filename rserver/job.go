package rserver

import (
	"errors"
	"fmt"
	"muskex/gen/mproto/model"
	"muskex/utils"
)

/*
php code:
public function firstRechargeReachedGive(Job $job, $data): void

	{
	    $userId = $data['user_id'];
	    $amount = $data['amount'];
	    $rechargeCount = CoinChange::where('user_id', $userId)->where('type', 'in', ['financial_recharge'])->count();
	    $rechargeGiveCount = CoinChange::where(['user_id' => $userId, 'type' => 'first_recharge_reached_give'])->count();
	    if ($rechargeCount > 1 || $rechargeGiveCount > 0) {
	        $job->delete();
	        return;
	    }
	    $taskGiveCoinType = get_sys_config('task_give_coin_type');
	    $giveArray = get_sys_config('first_recharge_reached_give');
	    array_multisort(array_column($giveArray,'key'),SORT_DESC, $giveArray);
	    foreach ($giveArray as $give) {
	        if ($amount >= $give['key'] && $give['value'] > 0) {
	            Assets::updateCoinAssetsBalance($userId, $taskGiveCoinType, $give['value'], 'first_recharge_reached_give', null, $give['key']);
	            break;
	        }
	    }
	    $job->delete();
	}
*/
func firstRechargeReachedGive(userId int64, amount float64) {
	var rechargeCount, rechargeGiveCount int64
	utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND type IN (?)", userId, []string{"financial_recharge"}).Count(&rechargeCount)
	utils.Orm.Model(&model.UserCoinChange{}).Where("user_id = ? AND type = ?", userId, "first_recharge_reached_give").Count(&rechargeGiveCount)
	if rechargeCount > 1 || rechargeGiveCount > 0 {
		return
	}
	taskGiveCoinType := int64(1) // Confirmed to only give U coins
	grades := collectGrades{}
	grades.Process("first_recharge_reached_give", amount, func(key, value float64) {
		if value > 0 {
			UpdateCoinAssetsBalance(utils.Orm, userId, taskGiveCoinType, value, "first_recharge_reached_give", 0, 0, fmt.Sprint(key))
		}
	})
}

// 对应user.php managementBuy
/*php code
  public static function updateCommission($userId, $amount, $type, $fromUserId = null, $remark = null)
  {
      $user = User::find($userId);
      if ($amount < 0 && $user->commission_pool < abs($amount)) {
          throw new Exception('佣金池余额不足');
      }

      $before = $user->commission_pool;
      $user->commission_pool += $amount;
      $user->save();

      // 保存佣金账变记录
      $commissionChange = [
          'user_id' => $userId,
          'amount' => $amount,
          'before' => $before,
          'after' => $user->commission_pool,
          'type' => $type,
          'from_user_id' => $fromUserId,
          'remark' => $remark,
      ];
      CommissionChange::create($commissionChange);
  }*/
func updateCommission(user *model.User, amount float64, changeType, remark string, orderid int64) error {
	if !(user.Refereeid != 0 && amount > 0) {
		return nil
	}
	if amount < 0 && user.CommissionPool < -amount {
		return errors.New("佣金池余额不足")
	}
	before := user.CommissionPool
	user.CommissionPool += amount
	tx := utils.Orm.Begin()
	defer tx.Rollback()
	if err := tx.Save(user).Error; err != nil {
		return err
	}
	commissionChange := &model.UserCommissionChange{
		UserId:     user.Id,
		Amount:     amount,
		Before:     before,
		After:      user.CommissionPool,
		Type:       changeType,
		FromUserId: user.Id,
		Remark:     remark,
		ReferrerId: orderid,
	}
	if err := tx.Create(commissionChange).Error; err != nil {
		return err
	}
	tx.Commit()
	return nil
}
