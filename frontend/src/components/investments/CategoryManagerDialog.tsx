import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import type { ChangeEvent, FormEvent } from 'react';

type Props = {
  open: boolean;
  saving: boolean;
  error: string | null;
  categories: string[];
  categoryDraft: string;
  editingCategoryName: string | null;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  setCategoryDraft: (value: string) => void;
  clearError: () => void;
  onStartEdit: (category: string) => void;
  onCancelEdit: () => void;
  onDelete: (category: string) => void;
};

export default function CategoryManagerDialog({
  open,
  saving,
  error,
  categories,
  categoryDraft,
  editingCategoryName,
  onClose,
  onSubmit,
  setCategoryDraft,
  clearError,
  onStartEdit,
  onCancelEdit,
  onDelete,
}: Props) {
  return (
    <Dialog open={open} onClose={saving ? undefined : onClose} fullWidth maxWidth="sm">
      <DialogTitle>Manage Categories</DialogTitle>
      <DialogContent dividers>
        <Stack component="form" spacing={2} onSubmit={onSubmit} id="category-form">
          {error ? <Alert severity="error">{error}</Alert> : null}

          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <TextField
              label={editingCategoryName === null ? 'New category' : 'Category name'}
              value={categoryDraft}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                setCategoryDraft(event.target.value);
                clearError();
              }}
              fullWidth
              required
            />
            <Button type="submit" variant="contained" disabled={saving}>
              {editingCategoryName === null ? 'Add Category' : 'Save Category'}
            </Button>
          </Stack>

          {editingCategoryName !== null ? (
            <Button onClick={onCancelEdit} disabled={saving} sx={{ alignSelf: 'flex-start' }}>
              Cancel Edit
            </Button>
          ) : null}

          <Typography variant="subtitle2" color="text.secondary">
            Uncategorized stays available automatically.
          </Typography>

          <Stack spacing={1}>
            {categories.length === 0 ? (
              <Typography color="text.secondary">No custom categories yet.</Typography>
            ) : (
              categories.map((category: string) => (
                <Stack key={category} direction="row" spacing={1} justifyContent="space-between" alignItems="center">
                  <Typography>{category}</Typography>
                  <Stack direction="row" spacing={1}>
                    <Button size="small" onClick={() => onStartEdit(category)} disabled={saving}>
                      Edit
                    </Button>
                    <Button size="small" color="error" onClick={() => onDelete(category)} disabled={saving}>
                      Delete
                    </Button>
                  </Stack>
                </Stack>
              ))
            )}
          </Stack>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={saving}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
