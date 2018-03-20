# Documentation for SQL Tables

## Events

| Column    | Type                     | Modifiers |
| --------- | ------------------------ | --------- |
| key       | text                     | not null  |
| name      | text                     | not null  |
| shortname | text                     |           |
| date      | timestamp with time zone | not null  |
| eventtype | integer                  | not null  |
| lat       | real                     |           |
| long      | real                     |           |
| enddate   | timestamp with time zone | not null  |

## Matches

| Column        | Type                     | Modifiers |
| ------------- | ------------------------ | --------- |
| key           | text                     | not null  |
| eventkey      | text                     | not null  |
| predictedtime | timestamp with time zone |           |
| actualtime    | timestamp with time zone |           |
| redscore      | integer                  |           |
| bluescore     | integer                  |           |
| youtubeurl    | text                     |           |

## Alliances

| Column   | Type    | Modifiers |
| -------- | ------- | --------- |
| matchkey | text    | not null  |
| isblue   | boolean | not null  |
| number   | text    | not null  |

## Photos

| Column | Type | Modifiers |
| ------ | ---- | --------- |
| team   | text | not null  |
| url    | text | not null  |

## Users

| Column         | Type    | Modifiers              |
| -------------- | ------- | ---------------------- |
| username       | text    | not null               |
| hashedpassword | text    | not null               |
| isadmin        | boolean | not null default false |

## Reports

| Column   | Type | Modifiers |
| -------- | ---- | --------- |
| reporter | text | not null  |
| eventkey | text | not null  |
| matchkey | text | not null  |
| team     | text | not null  |
| stats    | text | not null  |
| notes    | text |           |

## Picklists

| Column   | Type    | Collation | Nullable | Default                               |
| -------- | ------- | --------- | -------- | ------------------------------------- |
| id       | integer |           | not null | nextval('picklists_id_seq'::regclass) |
| eventkey | text    |           | not null |                                       |
| name     | text    |           |          |                                       |
| owner    | text    |           |          |                                       |

## Picks

| Column     | Type    | Collation | Nullable |
| ---------- | ------- | --------- | -------- |
| picklistid | integer |           | not null |
| team       | text    |           | not null |
