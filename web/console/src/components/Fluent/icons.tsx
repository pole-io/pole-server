import React from 'react';
import {
  Add20Regular,
  Apps20Regular,
  ArrowClockwise20Regular,
  ArrowRight20Regular,
  ArrowUndo20Regular,
  Building20Regular,
  Bot20Regular,
  Checkmark20Regular,
  CheckmarkCircle20Filled,
  CheckmarkCircle20Regular,
  ChevronDown20Regular,
  ChevronRight20Regular,
  Copy20Regular,
  Delete20Regular,
  Dismiss20Regular,
  DocumentAdd20Regular,
  DocumentCopy20Regular,
  Drag20Regular,
  Edit20Regular,
  Eye20Regular,
  EyeOff20Regular,
  Filter20Regular,
  Grid20Regular,
  Info20Regular,
  Layer20Regular,
  Link20Regular,
  PlugConnected20Regular,
  List20Regular,
  LockClosed20Regular,
  Mail20Regular,
  MoreHorizontal20Regular,
  Payment20Regular,
  People20Regular,
  Person20Regular,
  PersonAvailable20Regular,
  PersonCircle20Regular,
  Power20Regular,
  Pulse20Regular,
  QuestionCircle20Regular,
  Rocket20Regular,
  Save20Regular,
  Search20Regular,
  Send20Regular,
  Server20Regular,
  Settings20Regular,
  SignOut20Regular,
  SpinnerIos20Regular,
  Subtract20Regular,
  Tag20Regular,
  TextBulletListSquare20Regular,
  TextFont20Regular,
  WindowConsole20Regular,
  Wrench20Regular,
} from '@fluentui/react-icons';

export const AddIcon = Add20Regular;
export const RefreshIcon = ArrowClockwise20Regular;
export const DeleteIcon = Delete20Regular;
export const Delete1Icon = Delete20Regular;
export const Edit1Icon = Edit20Regular;
export const EditIcon = Edit20Regular;
export const RollbackIcon = ArrowUndo20Regular;
export const RocketIcon = Rocket20Regular;
export const CloseIcon = Dismiss20Regular;
export const ChevronRightIcon = ChevronRight20Regular;
export const SaveIcon = Save20Regular;
export const CreditcardIcon = Payment20Regular;
export const SearchIcon = Search20Regular;
export const BrowseIcon = Eye20Regular;
export const BrowseOffIcon = EyeOff20Regular;
export const LockOnIcon = LockClosed20Regular;
export const SendIcon = Send20Regular;
export const CopyIcon = Copy20Regular;
export const ServerIcon = Server20Regular;
export const UserIcon = Person20Regular;
export const User1Icon = Person20Regular;
export const UserCircleIcon = PersonCircle20Regular;
export const HelpCircleIcon = QuestionCircle20Regular;
export const SettingIcon = Settings20Regular;
export const ViewListIcon = List20Regular;
export const ListIcon = List20Regular;
export const LoadingIcon = SpinnerIos20Regular;
export const UsergroupIcon = People20Regular;
export const QueueIcon = TextBulletListSquare20Regular;
export const CheckCircleIcon = CheckmarkCircle20Regular;
export const CheckCircleFilledIcon = CheckmarkCircle20Filled;
export const IndicatorIcon = Pulse20Regular;
export const ViewModuleIcon = Grid20Regular;
export const AlphaIcon = TextFont20Regular;
export const LogoutIcon = SignOut20Regular;
export const ComponentSpaceIcon = Apps20Regular;
export const LayersIcon = Layer20Regular;
export const PoweroffIcon = Power20Regular;
export const MailIcon = Mail20Regular;
export const System2Icon = WindowConsole20Regular;
export const UserVisibleIcon = PersonAvailable20Regular;
export const DragMoveIcon = Drag20Regular;
export const FileCopyIcon = DocumentCopy20Regular;
export const DiscountIcon = Tag20Regular;
export const RemoveIcon = Subtract20Regular;
export const FileAddIcon = DocumentAdd20Regular;
export const ArrowRightIcon = ArrowRight20Regular;
export const ServiceIcon = Building20Regular;
export const CheckIcon = Checkmark20Regular;
export const FilterIcon = Filter20Regular;
export const TagIcon = Tag20Regular;
export const InfoCircleIcon = Info20Regular;
export const ToolsCircleIcon = Wrench20Regular;
export const LinkIcon = Link20Regular;
export const BotIcon = Bot20Regular;
export const PlugConnectedIcon = PlugConnected20Regular;

const iconByName: Record<string, React.ComponentType<any>> = {
  'user-circle': PersonCircle20Regular,
  'chevron-down': ChevronDown20Regular,
  more: MoreHorizontal20Regular,
  add: Add20Regular,
  delete: Delete20Regular,
  edit: Edit20Regular,
  refresh: ArrowClockwise20Regular,
  file: DocumentCopy20Regular,
  folder: Apps20Regular,
  'folder-open': Apps20Regular,
  code: WindowConsole20Regular,
  server: Server20Regular,
  service: Building20Regular,
};

interface IconProps extends React.SVGProps<SVGSVGElement> {
  name?: string;
}

export const Icon: React.FC<IconProps> = ({ name = '', ...props }) => {
  const Component = iconByName[name] || Apps20Regular;
  return <Component {...props} />;
};
