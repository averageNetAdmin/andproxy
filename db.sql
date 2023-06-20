BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS server ( 
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  address BLOB NOT NULL,
  port BLOB NOT NULL,
  -- timeout BIGINT,
  connect_timeout BIGINT,
  weight INT,
  max_fails BIGINT,
  break_time BIGINT,
  log_id BLOB,
  update_time BIGINT,
  create_time BIGINT,
  UNIQUE(address, port) ON CONFLICT FAIL
);
CREATE UNIQUE INDEX IF NOT EXISTS server_idx ON server (id);

CREATE TABLE IF NOT EXISTS pool ( 
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name BLOB NOT NULL UNIQUE,
  balancing_method BLOB NOT NULL,
  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB
);
-- CREATE UNIQUE INDEX server_idx ON server (id);

CREATE TABLE IF NOT EXISTS pool_servers ( 
  pool_id BIGINT,
  server_id BIGINT,
  disabled_server int DEFAULT 0,
  FOREIGN KEY (pool_id) REFERENCES pool(id),
  FOREIGN KEY (server_id) REFERENCES server(id)
);

CREATE TABLE IF NOT EXISTS acl ( 
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name BLOB NOT NULL UNIQUE,
  nets text NOT NULL,
  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB
);

CREATE TABLE IF NOT EXISTS proxy_filter (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name BLOB NOT NULL UNIQUE,
  target_net BLOB NOT NULL,
  server_pool_id BIGINT NOT NULL,
  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB,
  FOREIGN KEY (server_pool_id) REFERENCES pool(id)
);

CREATE TABLE IF NOT EXISTS proxy_filter_acls ( 
  filter_id BIGINT,
  acl_id BIGINT,
  deny_acl int DEFAULT 0,
  FOREIGN KEY (filter_id) REFERENCES proxy_filter(id),
  FOREIGN KEY (acl_id) REFERENCES acl(id)
);

CREATE TABLE IF NOT EXISTS l4handler ( 
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name BLOB NOT NULL UNIQUE,
  protocol BLOB NOT NULL,
  port BLOB NOT NULL,

  disabled INTEGER NOT NULL,
  deadline BIGINT NOT NULL,
  write_deadLine BIGINT NOT NULL,
  read_deadLine BIGINT NOT NULL,
  max_connections BIGINT,

  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB,
  UNIQUE(protocol, port) ON CONFLICT FAIL
);

CREATE TABLE IF NOT EXISTS http_handler ( 
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name BLOB NOT NULL UNIQUE,
  port BLOB NOT NULL,

  disabled INTEGER NOT NULL,
  secure INTEGER,
  max_connections BIGINT,

  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB,
  UNIQUE(port) ON CONFLICT FAIL
);

CREATE TABLE IF NOT EXISTS l4handler_proxy_filters ( 
  hanlder_id BIGINT NOT NULL,
  proxy_filter_id BIGINT NOT NULL,
  disabled_filter INT NOT NULL DEFAULT 0,
  FOREIGN KEY (hanlder_id) REFERENCES l4handler(id),
  FOREIGN KEY (proxy_filter_id) REFERENCES proxy_filter(id)
);

CREATE TABLE IF NOT EXISTS http_handler_http_hosts ( 
  host_id BIGINT NOT NULL,
  handler_id BIGINT NOT NULL,
  FOREIGN KEY (host_id) REFERENCES http_host(id),
  FOREIGN KEY (handler_id) REFERENCES http_handler(id)
);

CREATE TABLE IF NOT EXISTS http_host (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  address BLOB NOT NULL,
  cert_path BLOB NOT NULL,
  cert_key_path BLOB NOT NULL,
  max_connections BIGINT,
  disabled INT DEFAULT 0,

  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB
);

CREATE TABLE IF NOT EXISTS http_host_http_paths ( 
  host_id BIGINT NOT NULL,
  path_id BIGINT NOT NULL,
  FOREIGN KEY (host_id) REFERENCES http_host(id),
  FOREIGN KEY (path_id) REFERENCES http_path(id)
);

CREATE TABLE IF NOT EXISTS http_path (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path BLOB NOT NULL,
  enable_caching INT DEFAULT 1,
  request_timeout BIGINT NOT NULL,
  max_connections BIGINT,
  set_input BLOB,
  set_output BLOB,
  statuc_response BLOB,
  disabled INT DEFAULT 0,
  type BLOB NOT NULL,
  update_time BIGINT,
  create_time BIGINT,
  log_id BLOB
);

CREATE TABLE IF NOT EXISTS http_path_proxy_filters ( 
  path_id BIGINT NOT NULL,
  proxy_filter_id BIGINT NOT NULL,
  FOREIGN KEY (path_id) REFERENCES http_path(id),
  FOREIGN KEY (proxy_filter_id) REFERENCES proxy_filter(id)
);

CREATE TABLE IF NOT EXISTS logs ( 
  log_id BLOB,
  level BLOB NOT NULL,
  create_time BIGINT,
  message BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS http_cache (
  host_id INTEGER NOT NULL,
  path_id INTEGER NOT NULL,
  update_time BIGINT,
  create_time BIGINT,
  data BLOB NOT NULL
);

COMMIT;

-- docker run --mount type=bind,source="$(pwd)",target=/home/schcrwlr --rm -it schemacrawler/schemacrawler /opt/schemacrawler/bin/schemacrawler.sh --server=sqlite --database=db.sql --info-level=standard --command=schema --output-file=output.png