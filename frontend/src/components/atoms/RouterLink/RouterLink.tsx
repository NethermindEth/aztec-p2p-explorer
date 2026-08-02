import { Link, LinkProps } from 'react-router-dom';
import './RouterLink.css';

const RouterLink: React.FC<LinkProps> = ({ children, ...rest }) => {
  return <Link {...rest}>{children}</Link>;
};

export default RouterLink;
