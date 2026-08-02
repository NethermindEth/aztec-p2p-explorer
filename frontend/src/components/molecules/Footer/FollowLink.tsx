const FollowLink = ({ href, text, Icon }: { href: string; text: string; Icon: React.FC }) => (
  <a href={href} target="_blank" rel="noopener noreferrer" className="footer-follow">
    <div className="footer-follow-text">{text}</div>
    <div className="footer-follow-icon">
      <Icon />
    </div>
  </a>
);

export default FollowLink;
