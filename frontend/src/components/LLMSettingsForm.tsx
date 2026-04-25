import { useEffect, useMemo, useState, type ChangeEvent } from 'react';
import {
  Alert,
  Button,
  CircularProgress,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import type { LLMSettings, Provider } from '../lib/api';
import { deleteSettings, getSettings, saveSettings } from '../lib/api';

type Props = {
  provider: Provider;
  title: string;
  description: string;
};

const labelMap: Record<Provider, string> = {
  'lm-studio': 'LM Studio',
  ollama: 'Ollama',
  'external-endpoint': 'External Endpoint',
};

const emptyState = (provider: Provider): LLMSettings => ({
  provider,
  endpoint: '',
  model: '',
  apiKey: '',
  temperature: 0.7,
  systemPrompt: '',
});

export default function LLMSettingsForm({ provider, title, description }: Props) {
  const [settings, setSettings] = useState<LLMSettings>(emptyState(provider));
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const helper = useMemo(
    () => `These values persist in SQLite for ${labelMap[provider]}. Delete removes the saved row and the UI falls back to built-in defaults.`,
    [provider]
  );

  const load = async () => {
    setLoading(true);
    setError(null);
    setMessage(null);
    try {
      const result = await getSettings(provider);
      setSettings(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load settings');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, [provider]);

  const update = <K extends keyof LLMSettings>(key: K, value: LLMSettings[K]) => {
    setSettings((current: LLMSettings) => ({ ...current, [key]: value }));
  };

  const onSave = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      const result = await saveSettings(provider, settings);
      setSettings(result);
      setMessage('Settings saved.');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const onDelete = async () => {
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await deleteSettings(provider);
      setMessage('Saved record deleted. Reloaded default values.');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete settings');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <Stack spacing={2}>
        <Typography variant="h4">{title}</Typography>
        <Stack direction="row" spacing={2} alignItems="center">
          <CircularProgress size={24} />
          <Typography>Loading settings…</Typography>
        </Stack>
      </Stack>
    );
  }

  return (
    <Stack spacing={2}>
      <Typography variant="h4">{title}</Typography>
      <Typography color="text.secondary">{description}</Typography>

      {message ? <Alert severity="success">{message}</Alert> : null}
      {error ? <Alert severity="error">{error}</Alert> : null}

      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Typography variant="subtitle1">Connection settings</Typography>
          <TextField
            label="Endpoint"
            value={settings.endpoint}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('endpoint', event.target.value)}
            fullWidth
          />
          <TextField
            label="Model"
            value={settings.model}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('model', event.target.value)}
            fullWidth
          />
          <TextField
            label="API key"
            value={settings.apiKey}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('apiKey', event.target.value)}
            type="password"
            fullWidth
          />
          <TextField
            label="Temperature"
            value={settings.temperature}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('temperature', Number(event.target.value))}
            type="number"
            inputProps={{ min: 0, max: 2, step: 0.1 }}
            fullWidth
          />
          <TextField
            label="System prompt"
            value={settings.systemPrompt}
            onChange={(event: ChangeEvent<HTMLInputElement>) => update('systemPrompt', event.target.value)}
            multiline
            minRows={4}
            fullWidth
          />
          <Typography variant="body2" color="text.secondary">
            {helper}
          </Typography>
          <Stack direction="row" spacing={2}>
            <Button variant="contained" onClick={onSave} disabled={saving}>
              Save
            </Button>
            <Button variant="outlined" onClick={() => void load()} disabled={saving}>
              Reload
            </Button>
            <Button color="error" variant="outlined" onClick={onDelete} disabled={saving}>
              Delete saved record
            </Button>
          </Stack>
        </Stack>
      </Paper>
    </Stack>
  );
}
