import { useEffect, useState } from 'react';
import { Alert, Paper, Stack, Typography } from '@mui/material';
import { listSettings, type LLMSettings } from '../lib/api';

export default function Home() {
  const [rows, setRows] = useState<LLMSettings[]>([]);

  useEffect(() => {
    listSettings().then(setRows).catch(console.error);
  }, []);

  return (
    <Stack spacing={3}>
      <div>
        <Typography variant="h4" gutterBottom>
          Home
        </Typography>
        <Typography color="text.secondary">
          This starter includes an embedded Go server, SQLite-backed LLM settings, and a navigation layout ready for expansion.
        </Typography>
      </div>

      <Alert severity="info">
        LLM settings are stored in <strong>app-template.db</strong> beside the executable. When no saved record exists, the API returns provider-specific defaults.
      </Alert>

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="h6" gutterBottom>
          Effective LLM settings
        </Typography>
        <Stack spacing={1}>
          {rows.map((row: LLMSettings) => (
            <Typography key={row.provider}>
              <strong>{row.provider}</strong>: {row.model} @ {row.endpoint}
            </Typography>
          ))}
        </Stack>
      </Paper>
    </Stack>
  );
}
