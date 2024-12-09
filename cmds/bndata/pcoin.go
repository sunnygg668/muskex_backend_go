package main

import (
	"log"
	"muskex/utils"
)

// 所有平台币种
type pCoin struct {
	Id uint64
	// 交易所ID
	Pid             int64 `gorm:"uniqueIndex:pcoin_uni"`
	RawId           int64
	Name            string `gorm:"uniqueIndex:pcoin_uni"`
	Alias           string
	KlineType       string
	Margin          float64
	IsHot           int64
	HomeRecommended bool
	UpdateTime      int64
}

type pconf struct {
	//platform id
	Pid   int
	Dbstr string
}

func AggCons(pdbs []string) {
	//utils.Orm.AutoMigrate(pCoin{}, mproto.RankItem{})
	//return

	/*for pid, dbUrl := range pdbs {
			pid++
			lastTime := 0
			utils.Orm.Raw("select IFNULL(max(update_time),0) from ba_p_coin where pid=?", pid).Scan(&lastTime)

			db := utils.ConnectDb(dbUrl)
			sql := fmt.Sprintf(`
	select %d pid,t.id raw_id,
	       t.name,
	       t.alias,
	       t.kline_type,
	       t.margin,
	       t.is_hot,
	       t.home_recommend,
	       t.update_time
	from ba_coin t where t.status=1 and update_time>?;`, pid)
			list := []*pCoin{}
			err := db.Raw(sql, lastTime).Scan(&list).Error
			if err != nil {
				panic(err)
			}
			sqldb, _ := db.DB()
			if sqldb != nil {
				sqldb.Close()
			}

			if len(list) > 0 {
				log.Printf("AggCons pid: %d ,save RowsAffected %d", pid, utils.Orm.Save(list).RowsAffected)
			} else {
				log.Printf("AggCons pid: %d , no new data ", pid)
			}
		}*/
	//添加新币
	sql1 := `
insert into ba_rank_item(name, alias, kline_type)
select *
from (select distinct t.name,
                      t.alias,
                      t.kline_type
      from ba_p_coin t
      where not exists(select 1 from ba_rank_item where ba_rank_item.name = t.name)) abc;`
	newCount := utils.Orm.Exec(sql1).RowsAffected
	if newCount > 0 {
		log.Println("AggCons tableClean create new rank item count:", newCount)

	} else {
		log.Println("AggCons tableClean create none new rank item")
	}

	return

	//册除已删除的币
	sql2 := `
delete
from ba_rank_item
where not exists(select 1 from ba_p_coin where ba_p_coin.name = ba_rank_item.name);`
	delCount := utils.Orm.Exec(sql2).RowsAffected
	log.Println("AggCons delete rank item count:", delCount)

}
