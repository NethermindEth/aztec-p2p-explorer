import React from 'react';
import { XIcon, DiscordIcon } from '../../atoms/Icons/FooterIcons/FooterIcons';
import FollowLink from './FollowLink';
import CallToAction from './CallToAction';
import { useAtom } from 'jotai';
import { isMobileAtom } from '../../../hooks/atoms';
import NethermindLogo from '../../../assets/logos/nethermindLogo.svg';
import './Footer.css';

const callToActionItems = [
  {
    text: 'Run an Aztec Node',
    link: 'https://docs.aztec.network/operate/operators',
  },
];

const nethermindUrl = 'https://nethermind.io/';
const copyrightStatement = '©2025 Nethermind. All rights reserved.';
const aztecTwitter = 'https://x.com/aztecnetwork';
const aztecDiscord = 'https://discord.com/invite/aztec';

const footerFollowLinks = [
  { link: aztecTwitter, text: '', Icon: XIcon },
  { link: aztecDiscord, text: '', Icon: DiscordIcon },
];
const Footer: React.FC = () => {
  const [isMobile] = useAtom(isMobileAtom);
  return (
    <div className="footer">
      {isMobile ? (
        <div className="footer-mobile">
          <div className="footer-powered-by">
            <span>Powered by</span>
            <a href={nethermindUrl} target="_blank" rel="noopener noreferrer">
              <img src={NethermindLogo} alt="Nethermind Logo" />
            </a>
          </div>
          <div className="footer-rights">{copyrightStatement}</div>
          <div className="footer-mobile-middle">
            {callToActionItems.map((item) => (
              <CallToAction key={item.text} text={item.text} link={item.link} />
            ))}
            {footerFollowLinks.map((item, index) => (
              <FollowLink key={index} href={item.link} text={item.text} Icon={item.Icon} />
            ))}
          </div>
        </div>
      ) : (
        <>
          <div className="footer-left">
            <div className="footer-column">
              <div className="footer-column-content">
                <div className="footer-powered-by">
                  <span>Powered by</span>
                  <a href={nethermindUrl} target="_blank" rel="noopener noreferrer">
                    <img src={NethermindLogo} alt="Nethermind Logo" />
                  </a>
                </div>
                <div className="footer-rights">{copyrightStatement}</div>
              </div>
            </div>
          </div>
          <div className="footer-middle">
            {callToActionItems.map((item) => (
              <CallToAction key={item.text} text={item.text} link={item.link} />
            ))}
          </div>
          <div className="footer-right">
            {footerFollowLinks.map((item, index) => (
              <FollowLink key={index} href={item.link} text={item.text} Icon={item.Icon} />
            ))}
          </div>
        </>
      )}
    </div>
  );
};

export default Footer;
