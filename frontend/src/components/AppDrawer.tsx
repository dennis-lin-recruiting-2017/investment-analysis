import React, { type MouseEvent } from 'react';
import {
  Box,
  Button,
  Collapse,
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
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';
import { listInvestments, type Investment } from '../lib/api';

type NavItem = {
  label: string;
  path?: string;
  icon?: React.ReactNode;
  children?: NavItem[];
};

const navItems: NavItem[] = [
  {
    label: 'Home',
    path: '/home',
    icon: <HomeIcon />,
  },
  {
    label: 'Investments',
    path: '/investments/new',
    icon: <AccountBalanceWalletOutlinedIcon />,
    children: [],
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
  const [openGroups, setOpenGroups] = React.useState<Record<string, boolean>>({
    Investments: pathname.startsWith('/investments'),
    About: pathname.startsWith('/about'),
    'LLM Settings': pathname.startsWith('/settings'),
  });

  React.useEffect(() => {
    setOpenGroups((prev: Record<string, boolean>) => ({
      ...prev,
      Investments: prev.Investments || pathname.startsWith('/investments'),
      About: prev.About || pathname.startsWith('/about'),
      'LLM Settings': prev['LLM Settings'] || pathname.startsWith('/settings'),
    }));
  }, [pathname]);

  React.useEffect(() => {
    listInvestments()
      .then(setInvestments)
      .catch((err: unknown) => {
        console.error(err);
        setInvestments([]);
      });
  }, [pathname]);

  const toggleGroup = (label: string) => {
    setOpenGroups((prev: Record<string, boolean>) => ({ ...prev, [label]: !prev[label] }));
  };

  const drawerContent = (
    <Box>
      <Toolbar>
        <Typography variant="h6" noWrap component="div">
          App Template
        </Typography>
      </Toolbar>
      <Divider />
      <List disablePadding>
        {navItems.map((item) => {
          if (!item.children) {
            const selected = pathname === item.path;
            return (
              <ListItemButton
                key={item.label}
                selected={selected}
                onClick={() => onNavigate(item.path!)}
              >
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.label} />
              </ListItemButton>
            );
          }

          const children = item.label === 'Investments'
            ? investments.map((investment: Investment) => ({
                label: investment.ticker ? `${investment.name} (${investment.ticker})` : investment.name,
                path: `/investments/${investment.uuid}`,
                icon: <DescriptionIcon />,
              }))
            : item.children;
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
                <ListItemText primary={item.label} />
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
                    return (
                      <ListItemButton
                        key={child.label}
                        sx={{ pl: 4 }}
                        selected={selected}
                        onClick={() => onNavigate(child.path!)}
                      >
                        <ListItemIcon>{child.icon}</ListItemIcon>
                        <ListItemText primary={child.label} />
                      </ListItemButton>
                    );
                  })}
                </List>
              </Collapse>
            </React.Fragment>
          );
        })}
      </List>
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
