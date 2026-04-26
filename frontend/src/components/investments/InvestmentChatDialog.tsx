import { useEffect, useMemo, useState, type FormEvent } from 'react';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import { recordAnswer } from '../../investment-chat/engine';
import type { ChatAnswerValue, ChatAnswers, ChatHistoryEntry, ChatScript, ChatStep } from '../../investment-chat/types';

type Props = {
  open: boolean;
  script: ChatScript;
  saving: boolean;
  error: string | null;
  onClose: () => void;
  onComplete: (summary: string) => void;
};

function initialValueForStep(step: ChatStep): string {
  if (step.input === 'yesno') {
    return 'yes';
  }
  if (step.input === 'choice' && step.choices?.length) {
    return step.choices[0].value;
  }
  return '';
}

function toAnswerValue(step: ChatStep, value: string): ChatAnswerValue {
  if (step.input === 'number') {
    return Number(value);
  }
  if (step.input === 'yesno') {
    return value === 'yes';
  }
  return value;
}

function labelForSubmit(step: ChatStep, script: ChatScript, draftValue: string, answers: ChatAnswers): string {
  if (step.submitLabel) {
    return step.submitLabel;
  }
  if (script.getNextStepId(step.id, toAnswerValue(step, draftValue), { ...answers, [step.id]: toAnswerValue(step, draftValue) }) === null) {
    return script.completeLabel || 'Finish';
  }
  return 'Next';
}

export default function InvestmentChatDialog({
  open,
  script,
  saving,
  error,
  onClose,
  onComplete,
}: Props) {
  const [answers, setAnswers] = useState<ChatAnswers>({});
  const [history, setHistory] = useState<ChatHistoryEntry[]>([]);
  const [currentStepId, setCurrentStepId] = useState(script.initialStepId);
  const currentStep = useMemo(() => script.resolveStep(currentStepId, answers), [script, currentStepId, answers]);
  const [draftValue, setDraftValue] = useState(currentStep ? initialValueForStep(currentStep) : '');

  useEffect(() => {
    if (open) {
      setAnswers({});
      setHistory([]);
      setCurrentStepId(script.initialStepId);
    }
  }, [open, script]);

  useEffect(() => {
    if (currentStep) {
      setDraftValue(initialValueForStep(currentStep));
    }
  }, [currentStep]);

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!currentStep) {
      return;
    }

    const answerValue = toAnswerValue(currentStep, draftValue);
    const result = recordAnswer(script, currentStep, answerValue, answers, history);
    setAnswers(result.answers);
    setHistory(result.history);

    if (result.nextStepId === null) {
      onComplete(script.buildSummary(result.answers));
      return;
    }

    setCurrentStepId(result.nextStepId);
  };

  return (
    <Dialog
      open={open}
      onClose={saving ? undefined : onClose}
      fullWidth
      maxWidth="md"
      PaperProps={{
        sx: {
          maxHeight: '70vh',
        },
      }}
    >
      <DialogTitle>{script.title}</DialogTitle>
      <DialogContent dividers sx={{ overflowY: 'auto' }}>
        <Stack spacing={2}>
          {script.intro ? <Typography color="text.secondary">{script.intro}</Typography> : null}
          {error ? <Alert severity="error">{error}</Alert> : null}

          <Stack spacing={1}>
            {history.map((entry: ChatHistoryEntry) => (
              <Stack key={entry.stepId} spacing={1}>
                <Paper variant="outlined" sx={{ p: 2, alignSelf: 'flex-start', maxWidth: '85%' }}>
                  <Typography>{entry.prompt}</Typography>
                </Paper>
                <Paper variant="outlined" sx={{ p: 2, alignSelf: 'flex-end', maxWidth: '85%' }}>
                  <Typography>{entry.answerLabel}</Typography>
                </Paper>
              </Stack>
            ))}
          </Stack>

          {currentStep ? (
            <Paper variant="outlined" sx={{ p: 2 }}>
              <Stack component="form" spacing={2} onSubmit={handleSubmit}>
                <div>
                  <Typography>{currentStep.prompt}</Typography>
                  {currentStep.helpText ? (
                    <Typography variant="body2" color="text.secondary">
                      {currentStep.helpText}
                    </Typography>
                  ) : null}
                </div>

                {currentStep.input === 'textarea' ? (
                  <TextField
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    multiline
                    minRows={5}
                    fullWidth
                    autoFocus
                  />
                ) : null}

                {currentStep.input === 'text' ? (
                  <TextField
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    fullWidth
                    autoFocus
                  />
                ) : null}

                {currentStep.input === 'number' ? (
                  <TextField
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    type="number"
                    inputProps={{ step: 0.01 }}
                    fullWidth
                    autoFocus
                  />
                ) : null}

                {currentStep.input === 'date' ? (
                  <TextField
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    type="date"
                    InputLabelProps={{ shrink: true }}
                    fullWidth
                    autoFocus
                  />
                ) : null}

                {currentStep.input === 'choice' ? (
                  <TextField
                    select
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    fullWidth
                    autoFocus
                  >
                    {(currentStep.choices || []).map((choice) => (
                      <MenuItem key={choice.value} value={choice.value}>
                        {choice.label}
                      </MenuItem>
                    ))}
                  </TextField>
                ) : null}

                {currentStep.input === 'yesno' ? (
                  <TextField
                    select
                    value={draftValue}
                    onChange={(event) => setDraftValue(event.target.value)}
                    fullWidth
                    autoFocus
                  >
                    <MenuItem value="yes">Yes</MenuItem>
                    <MenuItem value="no">No</MenuItem>
                  </TextField>
                ) : null}

                <Box>
                  <Button type="submit" variant="contained" disabled={saving}>
                    {saving ? <CircularProgress size={20} color="inherit" /> : labelForSubmit(currentStep, script, draftValue, answers)}
                  </Button>
                </Box>
              </Stack>
            </Paper>
          ) : null}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Skip</Button>
      </DialogActions>
    </Dialog>
  );
}
