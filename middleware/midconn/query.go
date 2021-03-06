/**
 *  author: lim
 *  data  : 18-4-6 下午5:15
 */

package midconn

import (
	"github.com/lemonwx/xsql/sqlparser"
	"github.com/lemonwx/xsql/mysql"
	"github.com/lemonwx/log"
	"github.com/lemonwx/xsql/middleware/version"
	"time"
)

func (conn *MidConn) handleShow(stmt *sqlparser.Show, sql string) error {
	// show only send to one node
	rets, err := conn.ExecuteMultiNode(mysql.COM_QUERY, []byte(sql), []int{0,})
	if err != nil {
		log.Errorf("execute in multi node failed: %v", err)
		return err
	}

	return conn.HandleSelRets(rets)

}

func (conn *MidConn) handleSimpleSelect(stmt *sqlparser.SimpleSelect, sql string) error {
	rets, err := conn.ExecuteMultiNode(mysql.COM_QUERY, []byte(sql), nil)
	if err != nil {
		log.Errorf("execute in multi node failed: %v", err)
		return err
	}

	return conn.HandleSelRets(rets)
}

func (conn *MidConn) handleSelect(stmt *sqlparser.Select, sql string) error {

	hide := true

	ts := time.Now()
	var err error
	conn.VersionsInUse, err = version.VersionsInUse()
	if err != nil {
		log.Errorf("[%d] get VersionsInUse failed: %v", conn.ConnectionId, err)
		return err
	}
	log.Debugf("[%d] get VersionsInUse: %v", conn.ConnectionId, conn.VersionsInUse)


	if _, ok := stmt.SelectExprs[0].(*sqlparser.StarExpr); ok {
		log.Debugf("[%d] select * not need to convert", conn.ConnectionId)
	}

	if expr, ok := stmt.SelectExprs[0].(*sqlparser.NonStarExpr); ok {
		colName := sqlparser.String(expr)
		log.Debugf("[%d] select %s, expr, add extra col add first", conn.ConnectionId, colName)
		if colName != extraColName {
			tmp := make(sqlparser.SelectExprs, len(stmt.SelectExprs) + 1)
			copy(tmp[1:], stmt.SelectExprs[:])
			tmp[0] = &sqlparser.NonStarExpr{
				Expr:&sqlparser.ColName{Name: []byte(extraColName)},
			}
			stmt.SelectExprs = tmp
		} else {
			log.Debugf("[%d] select contains extra col, set hide = false", conn.ConnectionId)
			hide = false
		}
	}

	newSql := sqlparser.String(stmt)
	log.Debugf("[%d] convert sql to %s", conn.ConnectionId, newSql)

	if hide {
		conn.setupNodeStatus(conn.VersionsInUse, true)
		defer conn.setupNodeStatus(nil, false)
	}

	rets, err := conn.ExecuteMultiNode(mysql.COM_QUERY, []byte(newSql), nil)
	if err != nil {
		return err
	}

	err = conn.HandleSelRets(rets)
	log.Debugf("[%d] handle select cost: %v", conn.ConnectionId, time.Since(ts))

	return err
}

func (conn *MidConn) setupNodeStatus(vInUse [][]byte, hide bool) {
	for _, node := range conn.nodes {
		node.VersionsInUse = vInUse
		node.NeedHide = hide
	}
}