package discord

func dequeueUser(location string) (*JoinGuildQueue, error) {
	var item JoinGuildQueue

	// Begin a transaction
	tx := db.Begin()

	// Select the next item and lock it
	if err := tx.Raw(`SELECT * FROM queue_items WHERE status = 'pending' ORDER BY priority DESC, enqueue_time ASC LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&item).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete the item
	if err := tx.Delete(&QueueItem{}, item.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &item, nil
}
