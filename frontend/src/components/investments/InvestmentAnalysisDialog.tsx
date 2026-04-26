import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from '@mui/material';

type Props = {
  open: boolean;
  onClose: () => void;
};

export default function InvestmentAnalysisDialog({ open, onClose }: Props) {
  return (
    <Dialog
      open={open}
      onClose={onClose}
      fullWidth
      maxWidth="md"
      PaperProps={{
        sx: {
          maxHeight: '50vh',
        },
      }}
    >
      <DialogTitle>Investment Analysis</DialogTitle>
      <DialogContent
        dividers
        sx={{
          overflowY: 'auto',
        }}
      >
        <Stack spacing={2}>
          <Typography color="text.secondary">
            Investment analysis will appear here.
          </Typography>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
