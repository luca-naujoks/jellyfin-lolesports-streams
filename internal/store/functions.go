package store

import (
	"context"
	"database/sql"
	"fmt"
	"jellyfin-lolesports-streams/internal/types"
)

func (s *Store) GetAll() []types.Channel {
	var channels []types.Channel

	q := "SELECT id, name, provider, language, icon_url FROM channels"

	rows, err := s.db.Query(q)
	if err != nil {
		return nil
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("ERROR - %s", err.Error())
		}
	}(rows)

	for rows.Next() {
		var channel types.Channel

		err = rows.Scan(&channel.ID, &channel.Name, &channel.Provider, &channel.Language, &channel.IconUrl)
		if err != nil {
			return nil
		}

		leagues, err := getLeagues(s.db, channel.ID)
		if err != nil {
			return nil
		}

		channel.Leagues = leagues

		channels = append(channels, channel)
	}

	return channels
}

func (s *Store) GetByID(id string) (types.Channel, error) {
	var channel types.Channel

	channelQuery := "SELECT id, name, provider, language, icon_url FROM channels WHERE id = ?"

	row := s.db.QueryRow(channelQuery, id)
	err := row.Scan(&channel.ID, &channel.Name, &channel.Provider, &channel.Language, &channel.IconUrl)
	if err != nil {
		return types.Channel{}, err
	}

	leagues, err := getLeagues(s.db, id)

	channel.Leagues = leagues

	return channel, nil
}

func (s *Store) Add(channel types.Channel) error {
	addChannel := "INSERT INTO channels (id, name, provider, language, icon_url)VALUES (?,? ,? ,? ,? )"
	addLeague := "INSERT INTO leagues (id, slug, name, region, image)VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET slug = excluded.slug,name = excluded.name,region = excluded.region, image = excluded.image;"
	addChannelLeagueCon := "INSERT INTO channel_leagues (channel_id, league_id) VALUES (?,?) ON CONFLICT DO NOTHING"

	_, err := s.db.Exec(addChannel, channel.ID, channel.Name, channel.Provider, channel.Language, channel.IconUrl)
	if err != nil {
		return err
	}

	for _, league := range channel.Leagues {
		_, err = s.db.Exec(addLeague, league.Id, league.Slug, league.Name, league.Region, league.ImageUrl)
		if err != nil {
			return err
		}
		_, err = s.db.Exec(addChannelLeagueCon, channel.ID, league.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) Update(channel types.Channel) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			fmt.Printf("ERROR - %s", err.Error())
		}
	}(tx)

	updateChannel := "UPDATE channels SET name = ?, provider = ?, language = ?, icon_url = ? WHERE id = ?"
	_, err = tx.Exec(updateChannel, channel.Name, channel.Provider, channel.Language, channel.IconUrl, channel.ID)
	if err != nil {
		return err
	}

	rmLeagueLinks := "DELETE FROM channel_leagues WHERE channel_id = ?"
	_, err = tx.Exec(rmLeagueLinks, channel.ID)
	if err != nil {
		return err
	}

	addLeague := "INSERT INTO leagues (id, slug, name, region, image)VALUES (?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET slug = excluded.slug,name = excluded.name,region = excluded.region, image = excluded.image;"
	addChannelLeagueCon := "INSERT INTO channel_leagues (channel_id, league_id) VALUES (?,?) ON CONFLICT DO NOTHING"

	for _, league := range channel.Leagues {
		_, err = tx.Exec(addLeague, league.Id, league.Slug, league.Name, league.Region, league.ImageUrl)
		if err != nil {
			return err
		}
		_, err = tx.Exec(addChannelLeagueCon, channel.ID, league.Id)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(id string) error {
	deleteChannel := "DELETE FROM channels WHERE id = ?"
	_, err := s.db.Exec(deleteChannel, id)
	if err != nil {
		return err
	}

	return nil
}

func getLeagues(s *sql.DB, channelId string) ([]types.League, error) {
	var leagues []types.League

	channelLeagues := "SELECT l.id, l.slug, l.name, l.region, l.image FROM channel_leagues JOIN main.leagues l on channel_leagues.league_id = l.id WHERE channel_id = ?"
	rows, err := s.Query(channelLeagues, channelId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("ERROR - %s", err.Error())
		}
	}(rows)

	for rows.Next() {
		var league types.League
		err = rows.Scan(&league.Id, &league.Slug, &league.Name, &league.Region, &league.ImageUrl)
		if err != nil {
			return nil, err
		}

		leagues = append(leagues, league)
	}

	return leagues, nil
}
