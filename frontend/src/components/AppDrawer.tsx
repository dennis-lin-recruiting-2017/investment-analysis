import React, { type MouseEvent } from 'react';
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Collapse,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
} from '@mui/material';
import HomeIcon from '@mui/icons-material/Home';
import DescriptionIcon from '@mui/icons-material/Description';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import SettingsSuggestOutlinedIcon from '@mui/icons-material/SettingsSuggestOutlined';
import AccountBalanceWalletOutlinedIcon from '@mui/icons-material/AccountBalanceWalletOutlined';
import AddCircleOutlineIcon from '@mui/icons-material/AddCircleOutline';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';
import { deleteInvestment, listInvestments, type Investment } from '../lib/api';

type NavItem = {
  label: string;
  path?: string;
  icon?: React.ReactNode;
  children?: NavItem[];
};

type InvestmentNavItem = {
  investment: Investment;
  label: string;
  path: string;
  icon: React.ReactNode;
};

const navItems: NavItem[] = [
  {
    label: 'Home',
    path: '/home',
    icon: <HomeIcon />,
  },
  {
    label: 'About',
    path: '/about/author',
    icon: <InfoOutlinedIcon />,
    children: [
      { label: 'Author', path: '/about/author', icon: <DescriptionIcon /> },
      { label: 'Demo', path: '/about/demo', icon: <DescriptionIcon /> },
    ],
  },
  {
    label: 'LLM Settings',
    path: '/settings/lm-studio',
    icon: <SettingsSuggestOutlinedIcon />,
    children: [
      { label: 'LM Studio', path: '/settings/lm-studio', icon: <DescriptionIcon /> },
      { label: 'Ollama', path: '/settings/ollama', icon: <DescriptionIcon /> },
      { label: 'External Endpoint', path: '/settings/external-endpoint', icon: <DescriptionIcon /> },
    ],
  },
];

const investmentsNavItem: NavItem = {
  label: 'Investments',
  path: '/investments/new',
  icon: <AccountBalanceWalletOutlinedIcon />,
  children: [],
};

type AppDrawerProps = {
  drawerWidth: number;
  mobileOpen: boolean;
  onClose: () => void;
  pathname: string;
  onNavigate: (path: string) => void;
};

export default function AppDrawer({
  drawerWidth,
  mobileOpen,
  onClose,
  pathname,
  onNavigate,
}: AppDrawerProps) {
  const [investments, setInvestments] = React.useState<Investment[]>([]);
  const [deleteTarget, setDeleteTarget] = React.useState<Investment | null>(null);
  const [deleteError, setDeleteError] = React.useState<string | null>(null);
  const [deletingInvestment, setDeletingInvestment] = React.useState(false);
  const [openGroups, setOpenGroups] = React.useState<Record<string, boolean>>({
    Investments: pathname.startsWith('/investments'),
    About: pathname.startsWith('/about'),
    'LLM Settings': pathname.startsWith('/settings'),
  });

  const loadInvestments = React.useCallback(() => {
    listInvestments()
      .then(setInvestments)
      .catch((err: unknown) => {
        console.error(err);
        setInvestments([]);
      });
  }, []);

  React.useEffect(() => {
    setOpenGroups((prev: Record<string, boolean>) => ({
      ...prev,
      Investments: prev.Investments || pathname.startsWith('/investments'),
      About: prev.About || pathname.startsWith('/about'),
      'LLM Settings': prev['LLM Settings'] || pathname.startsWith('/settings'),
    }));
  }, [pathname]);

  React.useEffect(() => {
    loadInvestments();
  }, [loadInvestments, pathname]);

  const toggleGroup = (label: string) => {
    setOpenGroups((prev: Record<string, boolean>) => ({ ...prev, [label]: !prev[label] }));
  };

  const investmentChildren: InvestmentNavItem[] = investments.map((investment: Investment) => ({
    investment,
    label: investment.ticker ? `${investment.name} (${investment.ticker})` : investment.name,
    path: `/investments/${investment.uuid}`,
    icon: <DescriptionIcon />,
  }));

  const labelForItem = (item: NavItem): string => {
    if (item.label === 'Investments') {
      return `Investments (${investmentChildren.length})`;
    }
    return item.label;
  };

  const renderNavItem = (item: NavItem) => {
    if (!item.children) {
      const selected = pathname === item.path;
      return (
        <ListItemButton
          key={item.label}
          selected={selected}
          onClick={() => onNavigate(item.path!)}
        >
          <ListItemIcon>{item.icon}</ListItemIcon>
          <ListItemText primary={labelForItem(item)} />
        </ListItemButton>
      );
    }

    const children = item.label === 'Investments' ? investmentChildren : item.children;
    const isOpen = openGroups[item.label] ?? false;
    const parentSelected = pathname === item.path;
    const childSelected = children.some((child) => child.path === pathname);

    return (
      <React.Fragment key={item.label}>
        <ListItemButton
          component="div"
          selected={parentSelected || childSelected}
          onClick={() => toggleGroup(item.label)}
        >
          <ListItemIcon>{item.icon}</ListItemIcon>
          <ListItemText primary={labelForItem(item)} />
          {item.label === 'Investments' ? (
            <Button
              size="small"
              variant="outlined"
              startIcon={<AddCircleOutlineIcon />}
              onClick={(event: MouseEvent<HTMLButtonElement>) => {
                event.stopPropagation();
                onNavigate('/investments/new');
              }}
              sx={{ mr: 1 }}
            >
              New
            </Button>
          ) : null}
          <IconButton
            size="small"
            edge="end"
            onClick={(event: MouseEvent<HTMLButtonElement>) => {
              event.stopPropagation();
              toggleGroup(item.label);
            }}
          >
            {isOpen ? <ExpandLess /> : <ExpandMore />}
          </IconButton>
        </ListItemButton>

        <Collapse in={isOpen} timeout="auto" unmountOnExit>
          <List component="div" disablePadding>
            {item.label === 'Investments' && children.length === 0 ? (
              <ListItemText
                primary="No investments yet"
                sx={{ pl: 4, py: 1, color: 'text.secondary' }}
              />
            ) : null}
            {children.map((child) => {
              const selected = pathname === child.path;
              const investment = 'investment' in child ? child.investment : null;
              return (
                <ListItemButton
                  key={child.path || child.label}
                  sx={{ pl: 4, pr: 1 }}
                  selected={selected}
                  onClick={() => onNavigate(child.path!)}
                >
                  <ListItemIcon>{child.icon}</ListItemIcon>
                  <ListItemText primary={child.label} />
                  {investment ? (
                    <IconButton
                      size="small"
                      edge="end"
                      aria-label={`Delete ${child.label}`}
                      onClick={(event: MouseEvent<HTMLButtonElement>) => {
                        event.stopPropagation();
                        setDeleteError(null);
                        setDeleteTarget(investment);
                      }}
                    >
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  ) : null}
                </ListItemButton>
              );
            })}
          </List>
        </Collapse>
      </React.Fragment>
    );
  };

  const handleDeleteInvestment = async () => {
    if (!deleteTarget) {
      return;
    }

    setDeletingInvestment(true);
    setDeleteError(null);
    try {
      await deleteInvestment(deleteTarget.uuid);
      const deletedPath = `/investments/${deleteTarget.uuid}`;
      setDeleteTarget(null);
      await loadInvestments();
      if (pathname === deletedPath) {
        onNavigate('/home');
      }
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : 'Failed to delete investment');
    } finally {
      setDeletingInvestment(false);
    }
  };

  const drawerContent = (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <Toolbar>
        <Typography variant="h6" noWrap component="div">
          App Template
        </Typography>
      </Toolbar>
      <Divider />
      <List disablePadding>
        {navItems.map(renderNavItem)}
      </List>
      <Box sx={{ mt: 'auto' }}>
        <Divider />
        <List disablePadding>
          {renderNavItem(investmentsNavItem)}
        </List>
      </Box>

      <Dialog
        open={Boolean(deleteTarget)}
        onClose={deletingInvestment ? undefined : () => {
          setDeleteError(null);
          setDeleteTarget(null);
        }}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle>Delete Investment?</DialogTitle>
        <DialogContent dividers>
          <Box sx={{ display: 'grid', gap: 2 }}>
            {deleteError ? <Alert severity="error">{deleteError}</Alert> : null}
            <Typography>
              Delete {deleteTarget?.name || 'this investment'}? This will also remove its payment flows, sale assumptions, and categories.
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={() => {
              setDeleteError(null);
              setDeleteTarget(null);
            }}
            disabled={deletingInvestment}
          >
            Cancel
          </Button>
          <Button color="error" variant="contained" onClick={() => void handleDeleteInvestment()} disabled={deletingInvestment}>
            {deletingInvestment ? <CircularProgress size={20} color="inherit" /> : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );

  return (
    <Box component="nav" sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}>
      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={onClose}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: 'block', sm: 'none' },
          '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
        }}
      >
        {drawerContent}
      </Drawer>

      <Drawer
        variant="permanent"
        sx={{
          display: { xs: 'none', sm: 'block' },
          '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
        }}
        open
      >
        {drawerContent}
      </Drawer>
    </Box>
  );
}
