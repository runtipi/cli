#[cfg(test)]
mod test_is_major_bump {
    use crate::utils::release::is_major_bump;

    #[test]
    fn test_nightly() {
        assert!(!is_major_bump("1.2.3", "nightly"));
    }

    #[test]
    fn test_major_bump() {
        assert!(is_major_bump("1.2.3", "2.0.0"));
    }

    #[test]
    fn test_minor_bump() {
        assert!(!is_major_bump("1.2.3", "1.3.0"));
    }

    #[test]
    fn test_patch_bump() {
        assert!(!is_major_bump("1.2.3", "1.2.4"));
    }

    #[test]
    fn test_no_bump() {
        assert!(!is_major_bump("1.2.3", "1.2.3"));
    }

    #[test]
    fn test_downgrade() {
        assert!(!is_major_bump("2.0.0", "1.9.9"));
    }
}
