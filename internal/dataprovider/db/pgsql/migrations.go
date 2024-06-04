package pgsql

import (
	"github.com/forscht/ddrv/pkg/migrate"
)

var migrations = []migrate.Migration{
	{
		ID: 1,
		Up: migrate.Queries([]string{
			fsTable,
			nodeTable,
			fsParentIdx,
			fsNameIdx,
			rootInsert,
			statFunction,
			lsFunction,
			treeFunction,
			touchFunction,
			mkdirFunction,
			mvFunction,
			rmFunction,
			resetFunction,
			parserootFunction,
			validnameFunction,
			sanitizeFPath,
			parseSizeFunction,
			basenameFunction,
			dirnameFunction,
		}),
		Down: migrate.Queries([]string{dropFs}),
	},
	{
		ID: 2,
		Up: migrate.Queries([]string{
			`
				CREATE TABLE temp_node
				(
				    id    BIGINT PRIMARY KEY NOT NULL,
				    file  UUID               NOT NULL REFERENCES fs (id) ON DELETE CASCADE,
				    url   VARCHAR(255)       NOT NULL,
				    size  INTEGER            NOT NULL,
				    iv    VARCHAR(255)       NOT NULL DEFAULT '',
				    mtime TIMESTAMP          NOT NULL DEFAULT NOW()
				);
				
				INSERT INTO temp_node (id, file, url, size, iv, mtime)
				SELECT CAST(
				               (REGEXP_MATCHES(url, '/([0-9]+)/[A-Za-z0-9_-]+$', 'g'))[1]
				           AS BIGINT) AS id,
				       file,
				       url,
				       size,
				       iv,
				       mtime
				FROM node;
				
				DROP TABLE node;
				
				ALTER TABLE temp_node RENAME TO node;
				
				alter table public.node rename constraint temp_node_pkey to node_pkey;
				
				alter table public.node rename constraint temp_node_file_fkey to node_file_fkey;
			`,
		}),
		Down: migrate.Queries([]string{}),
	},
}
