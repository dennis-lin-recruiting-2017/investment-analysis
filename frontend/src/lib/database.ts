import initSqlJs, { type Database, type SqlJsStatic } from 'sql.js';

const STORAGE_KEY = 'app-template.sqlite';
let sqlJsPromise: Promise<SqlJsStatic> | undefined;
let databasePromise: Promise<Database> | undefined;

function getSqlJs(): Promise<SqlJsStatic> {
  if (sqlJsPromise === undefined) {
    sqlJsPromise = initSqlJs({
      locateFile: (file: string): string => `https://sql.js.org/dist/${file}`,
    });
  }

  return sqlJsPromise;
}

function loadPersistedDatabase(): Uint8Array | undefined {
  const saved = window.localStorage.getItem(STORAGE_KEY);
  if (!saved) {
    return undefined;
  }

  return Uint8Array.from(atob(saved), (char: string) => char.charCodeAt(0));
}

function persistDatabase(db: Database): void {
  const binaryArray = db.export();
  const binaryString = Array.from(binaryArray as Uint8Array)
    .map((byte: number) => String.fromCharCode(byte))
    .join('');
  window.localStorage.setItem(STORAGE_KEY, btoa(binaryString));
}

function seed(db: Database): void {
  db.run(`
    CREATE TABLE IF NOT EXISTS app_meta (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      key TEXT NOT NULL UNIQUE,
      value TEXT NOT NULL
    );
  `);

  const defaults: Array<[string, string]> = [
    ['storage_mode', 'browser-local'],
    ['db_file_name', 'app-template.db'],
    ['desktop_note', 'When this template is wrapped as a desktop executable, point the database path to the executable directory.'],
  ];

  defaults.forEach(([key, value]) => {
    db.run('INSERT OR IGNORE INTO app_meta (key, value) VALUES (?, ?)', [key, value]);
  });

  persistDatabase(db);
}

export async function getDatabase(): Promise<Database> {
  if (databasePromise === undefined) {
    databasePromise = getSqlJs().then((SQL: SqlJsStatic) => {
      const existing = loadPersistedDatabase();
      const db = existing ? new SQL.Database(existing) : new SQL.Database();
      seed(db);
      return db;
    });
  }

  return databasePromise;
}

export async function getAppMeta(): Promise<Array<{ key: string; value: string }>> {
  const db = await getDatabase();
  const result = db.exec('SELECT key, value FROM app_meta ORDER BY key');

  if (result.length === 0) {
    return [];
  }

  const values = result[0]?.values ?? [];
  return values.map((entry: unknown[]) => {
    const [key, value] = entry;
    return { key: String(key), value: String(value) };
  });
}

export async function upsertAppMeta(key: string, value: string): Promise<void> {
  const db = await getDatabase();
  db.run(
    `
      INSERT INTO app_meta (key, value)
      VALUES (?, ?)
      ON CONFLICT(key) DO UPDATE SET value = excluded.value
    `,
    [key, value]
  );
  persistDatabase(db);
}
