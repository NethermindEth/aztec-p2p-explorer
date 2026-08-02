import './Footer.css';
import { FooterIcon } from '../../atoms/Icons/FooterIcons/FooterIcons';

const CallToAction = ({ text, link }: { text: string; link: string }) => (
  <div className="footer-link">
    <a href={link} target="_blank" rel="noopener noreferrer" className="footer-link-text">
      {text}
    </a>
    <div className="footer-link-icon">
      <FooterIcon />
    </div>
  </div>
);

export default CallToAction;
