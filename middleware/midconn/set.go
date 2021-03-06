/**
 *  author: lim
 *  data  : 18-4-11 下午11:06
 */

package midconn

import (
	"github.com/lemonwx/xsql/sqlparser"
	"github.com/lemonwx/log"
	"github.com/lemonwx/xsql/mysql"
)

func (conn *MidConn)handleSet(stmt *sqlparser.Set, sql string) error {
	log.Debugf("[%d] handle set", conn.ConnectionId)
	if len(stmt.Exprs) != 1 {
		return UNEXPECT_MIDDLE_WARE_ERR
	}


	rets, err := conn.ExecuteMultiNode(mysql.COM_QUERY, []byte(sql), nil)
	if err != nil {
		return err
	}

	return conn.HandleExecRets(rets)

	expr := stmt.Exprs[0]

	if v, ok := expr.Expr.(sqlparser.NumVal); ok {
		log.Debugf("[%d], set num %v", conn.ConnectionId, v)
	}

	if v, ok := expr.Expr.(sqlparser.StrVal); ok {
		log.Debugf("[%d], set str d g%v", conn.ConnectionId, v)
	}

	/*
	if on :
		default = on
	if off
		default = off
	*/



	return conn.cli.WriteOK(nil)

}
