import React, { useContext } from 'react';
import { Link } from 'gatsby';
import { SharedContext } from '../state/shared';

export const Header = ({ changeLanguageCustomUrl }) => {
    const {
        language,
        getLanguagePrefix,
        setSearchVisible,
        setCartVisible,
        changeLanguage,
        currency,
        changeCurrency,
        cart,
    } = useContext(SharedContext);

    return (
        <header>
            <nav>
                <a
                    href='mailto:amadeus@pioneer.hr'
                    target='_blank'
                    rel='noopener noreferrer'
                >
                    amadeus@pioneer.hr
                </a>
                <a
                    href='tel:+38520670111'
                    target='_blank'
                    rel='noopener noreferrer'
                >
                    +385 (20) 670 111
                </a>
                {/* FIXME i18n */}
                <address>
                    <a
                        href='https://goo.gl/maps/74YCcjpq3PBWpKZK9'
                        target='_blank'
                        rel='noopener noreferrer'
                    >
                        Vladimira Nazora 45 Ploče, Hrvatska
                    </a>
                </address>
            </nav>
            <Link to={`${getLanguagePrefix(language)}/`}>Amadeus</Link>
            <button type='button' onClick={() => setSearchVisible(true)}>
                {/* FIXME add icon */}
                search
            </button>
            <button type='button' onClick={() => setCartVisible(true)}>
                {/* FIXME add icon */}
                cart ({cart.reduce((acc, obj) => acc + obj.quantity, 0)})
            </button>
            <select
                onChange={e =>
                    changeLanguage(e.target.value, changeLanguageCustomUrl)
                }
                value={language}
            >
                <option value='hr'>Hrvatski</option>
                <option value='en'>English</option>
            </select>
            <select
                value={currency}
                onChange={e => changeCurrency(e.target.value)}
            >
                <option value='HRK'>HRK</option>
                <option value='EUR'>EUR</option>
                <option value='BAM'>BAM</option>
                <option value='RSD'>RSD</option>
                <option value='USD'>USD</option>
                <option value='GBP'>GBP</option>
            </select>
        </header>
    );
};
