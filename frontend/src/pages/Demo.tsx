import { Paper, Stack, Typography } from '@mui/material';

export default function Demo() {
  return (
    <Stack spacing={2}>
      <Typography variant="h4">Demo</Typography>
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography paragraph>
          Use this page for demo walkthrough steps, screenshots, or quick-start instructions.
        </Typography>
        <Typography color="text.secondary">
          It is grouped under About in the left drawer.
        </Typography>
      </Paper>
    </Stack>
  );
}
