import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
} from '@mui/material';

type Props = {
  open: boolean;
  saving: boolean;
  categoryName: string | null;
  onClose: () => void;
  onConfirm: () => void;
};

export default function DeleteCategoryDialog({
  open,
  saving,
  categoryName,
  onClose,
  onConfirm,
}: Props) {
  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} maxWidth="xs" fullWidth>
      <DialogTitle>Delete Category?</DialogTitle>
      <DialogContent dividers>
        <Typography>
          Deleting <strong>{categoryName}</strong> will change all cash flows in that category to the
          {' '}Uncategorized option.
        </Typography>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Cancel</Button>
        <Button color="error" variant="contained" onClick={onConfirm} disabled={saving}>
          Delete Category
        </Button>
      </DialogActions>
    </Dialog>
  );
}
