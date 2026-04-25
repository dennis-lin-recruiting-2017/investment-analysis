import { Paper, Stack, Typography } from '@mui/material';

export default function Author() {
  return (
    <Stack spacing={2}>
      <Typography variant="h4">Author</Typography>
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography paragraph>
          Use this page for author details, credits, contact information, and project ownership notes.
        </Typography>
        <Typography color="text.secondary">
          It lives under About in the left drawer.
        </Typography>
      </Paper>
    </Stack>
  );
}
