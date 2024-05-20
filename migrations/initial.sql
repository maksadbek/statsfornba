create table player_avg (
    id serial primary key,
    name varchar(40) not null,
    team varchar(40) not null,
    season varchar(40) not null,
    points integer not null check (points >= 0),
	rebounds integer not null check (rebounds >= 0),
	assists integer not null check (assists >= 0),
	steals integer not null check (steals >= 0),
	blocks integer not null check (blocks >= 0),
	fouls integer not null check (fouls >= 0),
	turnovers integer not null check(turnovers >= 0),
	seconds_played integer not null check(seconds_played >= 0),
    game_count integer not null default 1
);

alter table player_avg add constraint player_avg_name_team_season unique (name, team, season);

create table team_avg (
    id serial primary key,
    name varchar(40) not null,
    season varchar(40) not null,
    points integer not null check (points >= 0),
	rebounds integer not null check (rebounds >= 0),
	assists integer not null check (assists >= 0),
	steals integer not null check (steals >= 0),
	blocks integer not null check (blocks >= 0),
	fouls integer not null check (fouls >= 0),
	turnovers integer not null check(turnovers >= 0),
	seconds_played integer not null check(seconds_played >= 0),
    game_count integer not null default 1
);

alter table team_avg add constraint team_avg_name_team_season unique (name, season);
