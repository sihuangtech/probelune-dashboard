package api

type historyBucket struct {
	NodeID       string   `json:"node_id"`
	At           int64    `json:"at"`
	Success      int64    `json:"success"`
	Failed       int64    `json:"failed"`
	Errors       int64    `json:"errors"`
	LatencySum   float64  `json:"latency_sum"`
	LatencyCount int64    `json:"latency_count"`
	Latency      *float64 `json:"latency_ms"`
}

func (s *Server) historyBuckets(monitor, node string, from, width int64) ([]historyBucket, error) {
	// 原始样本和已归档样本不重叠；相同桶合并，防止保留期边界重复计数。
	rows, err := s.Store.DB.Query(`SELECT node,bucket,SUM(success),SUM(failed),SUM(errors),SUM(latency_sum),SUM(latency_count) FROM (
 SELECT node,scheduled/?*? AS bucket,SUM(status='success') AS success,SUM(status='failure') AS failed,SUM(status NOT IN ('success','failure')) AS errors,COALESCE(SUM(latency),0) AS latency_sum,COUNT(latency) AS latency_count FROM runs WHERE monitor=? AND scheduled>=? AND (?='' OR node=?) GROUP BY node,scheduled/?
 UNION ALL SELECT node,bucket,success,failed,errors,latency_sum,latency_count FROM buckets WHERE monitor=? AND width=? AND bucket>=? AND (?='' OR node=?)
 ) GROUP BY node,bucket ORDER BY bucket`, width, width, monitor, from, node, node, width, monitor, width, from, node, node)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []historyBucket{}
	for rows.Next() {
		var b historyBucket
		if err = rows.Scan(&b.NodeID, &b.At, &b.Success, &b.Failed, &b.Errors, &b.LatencySum, &b.LatencyCount); err != nil {
			return nil, err
		}
		if b.LatencyCount > 0 {
			v := b.LatencySum / float64(b.LatencyCount)
			b.Latency = &v
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
